#!/usr/bin/env python3
"""
집계자 배치 최적화 스크립트
NSGA-II 알고리즘을 사용하여 멀티클라우드 환경에서 최적의 집계자 배치를 찾습니다.
"""

import json
import sys
import time
import requests
import subprocess
import statistics
import concurrent.futures
import threading
from typing import List, Dict, Tuple
import numpy as np
from deap import base, creator, tools, algorithms
import random

# 전역 상수
CANDIDATE_REGIONS = [
    {"region": "us-east-1", "provider": "AWS", "location": "Virginia"},
    {"region": "us-west-2", "provider": "AWS", "location": "Oregon"},
    {"region": "eu-west-1", "provider": "AWS", "location": "Ireland"},
    {"region": "ap-northeast-1", "provider": "AWS", "location": "Tokyo"},
    {"region": "ap-northeast-2", "provider": "AWS", "location": "Seoul"},
    {"region": "us-central1", "provider": "GCP", "location": "Iowa"},
    {"region": "europe-west1", "provider": "GCP", "location": "Belgium"},
    {"region": "asia-northeast1", "provider": "GCP", "location": "Tokyo"},
    {"region": "asia-northeast3", "provider": "GCP", "location": "Seoul"},
]

INSTANCE_TYPES = [
    {"type": "m1.small", "cpu": 1, "memory": 2, "provider": "openstack"},
    {"type": "m1.medium", "cpu": 2, "memory": 4, "provider": "openstack"},
    {"type": "m1.large", "cpu": 4, "memory": 8, "provider": "openstack"},
    {"type": "m1.xlarge", "cpu": 8, "memory": 16, "provider": "openstack"},
]

class RTTMeasurer:
    """RTT 측정을 담당하는 클래스"""
    
    def __init__(self, participants: List[Dict]):
        self.participants = participants
        self.rtt_cache = {}
        self.lock = threading.Lock()
    
    def measure_rtt_to_region(self, participant_endpoint: str, region: str) -> Dict[str, float]:
        """특정 참여자에서 특정 리전까지의 RTT 측정 (상세 통계 포함)"""
        cache_key = f"{participant_endpoint}_{region}"
        
        with self.lock:
            if cache_key in self.rtt_cache:
                return self.rtt_cache[cache_key]
        
        try:
            # 리전에 따른 대표 IP로 ping 측정
            target_ips = self._get_region_target_ip(region)
            
            best_result = None
            for target_ip in target_ips:
                result = self._ping_target_detailed(target_ip, count=5)
                if result and result['avg_latency'] > 0:
                    if best_result is None or result['avg_latency'] < best_result['avg_latency']:
                        best_result = result
            
            if best_result is None:
                # 측정 실패 시 지역별 예상 RTT 사용
                best_result = {
                    'min_latency': self._estimate_rtt_by_region(region),
                    'avg_latency': self._estimate_rtt_by_region(region),
                    'max_latency': self._estimate_rtt_by_region(region) * 1.2,
                    'std_dev': 5.0,
                    'packet_loss': 0.0
                }
            
            with self.lock:
                self.rtt_cache[cache_key] = best_result
            
            return best_result
            
        except Exception as e:
            print(f"RTT 측정 실패 ({participant_endpoint} -> {region}): {e}")
            # 실패 시 예상값 반환
            fallback_rtt = self._estimate_rtt_by_region(region)
            return {
                'min_latency': fallback_rtt,
                'avg_latency': fallback_rtt,
                'max_latency': fallback_rtt * 1.2,
                'std_dev': 5.0,
                'packet_loss': 0.0
            }
    
    def _get_region_target_ip(self, region: str) -> List[str]:
        """리전별 대표 IP 반환 (Go 코드 참고)"""
        region_targets = {
            # AWS 리전
            "us-east-1": ["13.231.217.241"],  # AWS 버지니아
            "us-west-2": ["52.77.229.124"],   # AWS 오레곤
            "eu-west-1": ["3.80.156.106"],    # AWS 아일랜드
            "ap-northeast-1": ["13.231.217.241"],  # AWS 도쿄
            "ap-northeast-2": ["52.79.173.168"],   # AWS 서울
            
            # GCP 리전 (예시 IP)
            "us-central1": ["8.8.8.8"],      # GCP 아이오와
            "europe-west1": ["8.8.4.4"],     # GCP 벨기에
            "asia-northeast1": ["13.231.217.241"],  # GCP 도쿄
            "asia-northeast3": ["52.79.173.168"],   # GCP 서울
        }
        return region_targets.get(region, ["8.8.8.8"])
    
    def _ping_target_detailed(self, target_ip: str, count: int = 5) -> Dict[str, float]:
        """상세한 ping 통계를 포함한 RTT 측정 (Go 코드 스타일 적용)"""
        try:
            # ping 명령어 실행 (Linux/Mac 호환)
            cmd = ["ping", "-c", str(count), "-W", "2000", target_ip]
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=15  # 전체 타임아웃
            )
            
            if result.returncode != 0:
                return None
            
            # ping 결과 파싱
            lines = result.stdout.split('\n')
            rtts = []
            packet_sent = count
            packet_received = 0
            
            # RTT 값들 추출
            for line in lines:
                if "time=" in line:
                    try:
                        time_part = line.split("time=")[1].split()[0]
                        rtt_value = float(time_part.replace("ms", ""))
                        rtts.append(rtt_value)
                        packet_received += 1
                    except (IndexError, ValueError):
                        continue
            
            if not rtts:
                return None
            
            # 통계 계산 (Go 코드와 동일한 방식)
            min_rtt = min(rtts)
            max_rtt = max(rtts)
            avg_rtt = statistics.mean(rtts)
            std_dev = statistics.stdev(rtts) if len(rtts) > 1 else 0.0
            packet_loss = ((packet_sent - packet_received) / packet_sent) * 100
            
            return {
                'min_latency': min_rtt,
                'avg_latency': avg_rtt,
                'max_latency': max_rtt,
                'std_dev': std_dev,
                'packet_loss': packet_loss
            }
            
        except Exception as e:
            print(f"Detailed ping 실패 ({target_ip}): {e}")
            return None
    
    def _estimate_rtt_by_region(self, region: str) -> float:
        """지역별 예상 RTT (측정 실패 시 사용)"""
        # 한국에서 각 리전까지의 대략적인 RTT (ms)
        estimated_rtts = {
            "us-east-1": 180,
            "us-west-2": 150,
            "eu-west-1": 280,
            "ap-northeast-1": 30,
            "ap-northeast-2": 10,
            "us-central1": 160,
            "europe-west1": 300,
            "asia-northeast1": 30,
            "asia-northeast3": 10,
        }
        return estimated_rtts.get(region, 100)
    
    def measure_all_rtts(self) -> Dict[str, Dict[str, float]]:
        """모든 참여자-리전 조합의 RTT 측정 (병렬 처리)"""
        rtt_matrix = {}
        
        print("RTT 측정 시작...")
        total_measurements = len(self.participants) * len(CANDIDATE_REGIONS)
        current = 0
        
        # 병렬 처리를 위한 작업 리스트 생성
        tasks = []
        for participant in self.participants:
            participant_id = participant['id']
            endpoint = participant['openstack_endpoint']
            rtt_matrix[participant_id] = {}
            
            for region_info in CANDIDATE_REGIONS:
                region = region_info['region']
                tasks.append((participant_id, endpoint, region))
        
        # 병렬 RTT 측정 실행 (Go 코드의 goroutine과 유사)
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            # 모든 작업 제출
            future_to_task = {
                executor.submit(self.measure_rtt_to_region, endpoint, region): (participant_id, region)
                for participant_id, endpoint, region in tasks
            }
            
            # 결과 수집
            for future in concurrent.futures.as_completed(future_to_task):
                participant_id, region = future_to_task[future]
                try:
                    result = future.result()
                    # 평균 RTT만 사용 (기존 코드와 호환성 유지)
                    rtt_matrix[participant_id][region] = result['avg_latency']
                    
                    current += 1
                    progress = (current / total_measurements) * 100
                    print(f"RTT 측정 진행률: {progress:.1f}% ({current}/{total_measurements}) - "
                          f"{participant_id} -> {region}: {result['avg_latency']:.2f}ms "
                          f"(손실: {result['packet_loss']:.1f}%)")
                    
                except Exception as e:
                    print(f"RTT 측정 작업 실패 ({participant_id} -> {region}): {e}")
                    # 실패 시 예상값 사용
                    rtt_matrix[participant_id][region] = self._estimate_rtt_by_region(region)
                    current += 1
        
        print("RTT 측정 완료!")
        
        # 측정 결과 요약 출력 (Go 코드 스타일)
        self._print_rtt_summary(rtt_matrix)
        
        return rtt_matrix
    
    def _print_rtt_summary(self, rtt_matrix: Dict[str, Dict[str, float]]):
        """RTT 측정 결과 요약 출력"""
        print("\n== RTT 측정 결과 요약 ==")
        
        for participant_id, rtts in rtt_matrix.items():
            print(f"\n참여자: {participant_id}")
            print("대상 리전              | 평균 Latency (ms)")
            print("-----------------------|------------------")
            
            for region, rtt in sorted(rtts.items()):
                print(f"{region:<22} | {rtt:>16.2f}")
        
        print("\n" + "="*50)

class CostCalculator:
    """비용 계산을 담당하는 클래스"""
    
    def __init__(self, cost_data_path: str = "./data/cloud_costs.json"):
        self.cost_data = self._load_cost_data(cost_data_path)
    
    def _load_cost_data(self, cost_data_path: str) -> Dict:
        """미리 파싱된 비용 정보 JSON 파일 로드"""
        try:
            with open(cost_data_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except FileNotFoundError:
            print(f"비용 데이터 파일을 찾을 수 없습니다: {cost_data_path}")
            # 기본 비용 데이터 사용
            return self._get_default_cost_data()
    
    def _get_default_cost_data(self) -> Dict:
        """기본 비용 데이터 (파일이 없을 경우 사용)"""
        return {
            "aws": {
                "us-east-1": {"m1.small": 0.05, "m1.medium": 0.10, "m1.large": 0.20, "m1.xlarge": 0.40},
                "us-west-2": {"m1.small": 0.055, "m1.medium": 0.11, "m1.large": 0.22, "m1.xlarge": 0.44},
                "eu-west-1": {"m1.small": 0.06, "m1.medium": 0.12, "m1.large": 0.24, "m1.xlarge": 0.48},
                "ap-northeast-1": {"m1.small": 0.058, "m1.medium": 0.116, "m1.large": 0.232, "m1.xlarge": 0.464},
                "ap-northeast-2": {"m1.small": 0.052, "m1.medium": 0.104, "m1.large": 0.208, "m1.xlarge": 0.416},
            },
            "gcp": {
                "us-central1": {"m1.small": 0.048, "m1.medium": 0.096, "m1.large": 0.192, "m1.xlarge": 0.384},
                "europe-west1": {"m1.small": 0.058, "m1.medium": 0.116, "m1.large": 0.232, "m1.xlarge": 0.464},
                "asia-northeast1": {"m1.small": 0.056, "m1.medium": 0.112, "m1.large": 0.224, "m1.xlarge": 0.448},
                "asia-northeast3": {"m1.small": 0.050, "m1.medium": 0.100, "m1.large": 0.200, "m1.xlarge": 0.400},
            }
        }
    
    def get_hourly_cost(self, region: str, instance_type: str) -> float:
        """시간당 비용 계산"""
        provider = self._get_provider_by_region(region)
        
        try:
            return self.cost_data[provider][region][instance_type]
        except KeyError:
            print(f"비용 정보를 찾을 수 없습니다: {provider}/{region}/{instance_type}")
            return 0.1  # 기본값
    
    def _get_provider_by_region(self, region: str) -> str:
        """리전명으로 클라우드 제공업체 판별"""
        aws_regions = ["us-east-1", "us-west-2", "eu-west-1", "ap-northeast-1", "ap-northeast-2"]
        return "aws" if region in aws_regions else "gcp"

class NSGAIIOptimizer:
    """NSGA-II 다목적 최적화 클래스"""
    
    def __init__(self, rtt_matrix: Dict, cost_calculator: CostCalculator, constraints: Dict):
        self.rtt_matrix = rtt_matrix
        self.cost_calculator = cost_calculator
        self.constraints = constraints
        self.aggregator_options = self._generate_aggregator_options()
        self.filtered_options = self._filter_options_by_constraints()
    
    def _generate_aggregator_options(self) -> List[Dict]:
        """모든 가능한 집계자 배치 옵션 생성"""
        options = []
        option_id = 0
        
        for region_info in CANDIDATE_REGIONS:
            for instance_info in INSTANCE_TYPES:
                region = region_info['region']
                instance_type = instance_info['type']
                
                # 비용 계산
                hourly_cost = self.cost_calculator.get_hourly_cost(region, instance_type)
                monthly_cost = hourly_cost * 24 * 30 * 1300  # 월 비용 (원)
                
                # 평균 지연시간 계산
                total_rtt = 0
                participant_count = 0
                
                for participant_id, rtts in self.rtt_matrix.items():
                    if region in rtts:
                        total_rtt += rtts[region]
                        participant_count += 1
                
                avg_latency = total_rtt / participant_count if participant_count > 0 else 1000
                
                option = {
                    'id': option_id,
                    'region': region,
                    'instanceType': instance_type,
                    'cloudProvider': region_info['provider'],
                    'totalCost': monthly_cost,
                    'totalRTT': avg_latency
                }
                options.append(option)
                option_id += 1
        
        return options
    
    def _filter_options_by_constraints(self) -> List[Dict]:
        """사용자 제약사항에 따른 옵션 필터링"""
        filtered = []
        
        for option in self.aggregator_options:
            if (option['totalCost'] <= self.constraints['maxBudget'] and 
                option['totalRTT'] <= self.constraints['maxLatency']):
                filtered.append(option)
        
        print(f"제약사항 필터링: {len(self.aggregator_options)} -> {len(filtered)} 옵션")
        
        if not filtered:
            raise ValueError("사용자 제약사항에 맞는 집계자 배치 옵션이 없습니다.")
        
        return filtered
    
    def optimize(self) -> List[Dict]:
        """NSGA-II 최적화 실행"""
        print("NSGA-II 최적화 시작...")
        
        if not self.filtered_options:
            raise ValueError("최적화할 옵션이 없습니다.")
        
        # DEAP 설정
        creator.create("FitnessMulti", base.Fitness, weights=(-1.0, -1.0))  # 비용, 지연시간 최소화
        creator.create("Individual", list, fitness=creator.FitnessMulti)
        
        toolbox = base.Toolbox()
        toolbox.register("attr_int", random.randint, 0, len(self.filtered_options) - 1)
        toolbox.register("individual", tools.initRepeat, creator.Individual, toolbox.attr_int, 1)
        
        population_size = max(10, len(self.filtered_options) // 10)  # 최소 10개 개체
        toolbox.register("population", tools.initRepeat, list, toolbox.individual, population_size)
        
        # 함수 등록
        toolbox.register("evaluate", self._evaluate_individual)
        toolbox.register("mate", tools.cxUniform, indpb=0.5)
        toolbox.register("mutate", tools.mutUniformInt, 
                        low=0, up=len(self.filtered_options) - 1, indpb=0.5)
        toolbox.register("select", tools.selNSGA2)
        
        # 초기 집단 생성
        population = toolbox.population()
        
        # 유전 알고리즘 파라미터
        ngen = 100  # 세대 수
        mu = population_size  # 부모 개체 수
        lambda_ = int(mu * 1.5)  # 자손 개체 수
        cxpb, mutpb = 0.8, 0.1  # 교차 및 돌연변이 확률
        
        # 진화 알고리즘 실행
        algorithms.eaMuPlusLambda(
            population, toolbox, mu=mu, lambda_=lambda_, 
            cxpb=cxpb, mutpb=mutpb, ngen=ngen,
            stats=None, halloffame=None, verbose=False
        )
        
        # Pareto 최적해 추출
        pareto_front = tools.sortNondominated(population, len(population), first_front_only=True)[0]
        
        # 중복 제거
        unique_options = set()
        final_pareto_front = []
        
        for individual in pareto_front:
            option_id = individual[0]
            option_key = (self.filtered_options[option_id]['region'], 
                         self.filtered_options[option_id]['instanceType'])
            
            if option_key not in unique_options:
                unique_options.add(option_key)
                final_pareto_front.append(individual)
        
        print(f"NSGA-II 최적화 완료! 파레토 최적해 {len(final_pareto_front)}개 발견")
        
        # 가중합을 사용한 최종 정렬
        return self._convert_to_results(final_pareto_front)
    
    def _evaluate_individual(self, individual) -> Tuple[float, float]:
        """개체 평가 함수"""
        option = self.filtered_options[individual[0]]
        return option['totalCost'], option['totalRTT']
    
    def _convert_to_results(self, pareto_front) -> List[Dict]:
        """파레토 최적해를 결과 형식으로 변환"""
        results = []
        
        for individual in pareto_front:
            option = self.filtered_options[individual[0]]
            
            result = {
                "rank": 0,  # 임시, 나중에 재배정
                "region": option['region'],
                "instanceType": option['instanceType'],
                "estimatedCost": round(option['totalCost'], 2),
                "estimatedLatency": round(option['totalRTT'], 2),
                "cloudProvider": option['cloudProvider']
            }
            results.append(result)
        
        # 가중합을 사용한 정렬 (비용 0.5, 지연시간 0.5)
        results = self._select_weighted_best_multiple(results)
        
        # 순위 재배정
        for i, result in enumerate(results):
            result['rank'] = i + 1
        
        return results
    
    def _select_weighted_best_multiple(self, results, rtt_weight=0.5, cost_weight=0.5) -> List[Dict]:
        """가중합을 사용하여 결과 정렬"""
        # 정규화를 위한 최대값 계산
        max_cost = max(result['estimatedCost'] for result in results) if results else 1
        max_latency = max(result['estimatedLatency'] for result in results) if results else 1
        
        # 가중합 점수 계산 및 정렬
        for result in results:
            normalized_cost = result['estimatedCost'] / max_cost
            normalized_latency = result['estimatedLatency'] / max_latency
            result['_score'] = cost_weight * normalized_cost + rtt_weight * normalized_latency
        
        # 점수 기준으로 정렬
        results.sort(key=lambda x: x['_score'])
        
        # 임시 점수 필드 제거
        for result in results:
            del result['_score']
        
        return results

def main():
    """메인 함수"""
    if len(sys.argv) != 3:
        print("사용법: python3 aggregator_optimization.py <input_file> <output_file>")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    try:
        # 1. 입력 파일 읽기
        with open(input_file, 'r', encoding='utf-8') as f:
            input_data = json.load(f)
        
        print(f"입력 데이터 로드 완료: {input_file}")
        
        # 2. 데이터 추출
        participants = input_data['federatedLearning']['participants']
        constraints = input_data['constraints']
        
        print(f"참여자 수: {len(participants)}")
        print(f"제약사항: 최대예산 {constraints['maxBudget']}원, 최대지연시간 {constraints['maxLatency']}ms")
        
        # 3. RTT 측정
        rtt_measurer = RTTMeasurer(participants)
        rtt_matrix = rtt_measurer.measure_all_rtts()
        
        # 4. 비용 계산기 초기화
        cost_calculator = CostCalculator()
        
        # 5. NSGA-II 최적화 실행
        optimizer = NSGAIIOptimizer(rtt_matrix, cost_calculator, constraints)
        optimization_results = optimizer.optimize()
        
        # 6. 결과 저장
        result = {
            "optimizationResults": optimization_results,
            "executionTime": 0,  # 실제 실행 시간은 Go에서 계산
            "status": "completed"
        }
        
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(result, f, indent=2, ensure_ascii=False)
        
        print(f"최적화 결과 저장 완료: {output_file}")
        print(f"최적해 개수: {len(optimization_results)}")
        
    except Exception as e:
        # 에러 발생 시 에러 정보를 출력 파일에 저장
        error_result = {
            "optimizationResults": [],
            "executionTime": 0,
            "status": "error",
            "error": str(e)
        }
        
        try:
            with open(output_file, 'w', encoding='utf-8') as f:
                json.dump(error_result, f, indent=2, ensure_ascii=False)
        except:
            pass
        
        print(f"오류 발생: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()