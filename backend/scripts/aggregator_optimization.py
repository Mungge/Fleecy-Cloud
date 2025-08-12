#!/usr/bin/env python3
"""
집계자 배치 최적화 API 스크립트 (PostgreSQL 연동)
프론트엔드 요청을 받아 NSGA-II 알고리즘으로 최적화된 집계자 옵션 리스트를 반환합니다.
"""

import json
import sys
import os
import psycopg2
from typing import List, Dict, Tuple, Optional
import numpy as np
from deap import base, creator, tools, algorithms
import random
from dotenv import load_dotenv

# 환경변수 로드
load_dotenv()

# 전역 상수
USD_TO_KRW = float(os.getenv('USD_TO_KRW', '1300'))

class DatabaseManager:
    """PostgreSQL 데이터베이스 연결 관리"""
    
    def __init__(self):
        self.conn = psycopg2.connect(
            host=os.getenv('DB_HOST'),
            user=os.getenv('DB_USER'), 
            password=os.getenv('DB_PASSWORD'),
            database=os.getenv('DB_NAME'),
            port=int(os.getenv('DB_PORT', '5432'))
        )
    
    def get_cloud_prices(self) -> List[Dict]:
        """클라우드 가격 정보 조회"""
        with self.conn.cursor() as cursor:
            cursor.execute("""
                SELECT cloud_name, region_name, instance_type, v_cpu_count, 
                       memory_gb, on_demand_price
                FROM cloud_price
                ORDER BY cloud_name, region_name, on_demand_price
            """)
            return [
                {
                    'cloud_name': row[0],
                    'region_name': row[1], 
                    'instance_type': row[2],
                    'v_cpu_count': row[3],
                    'memory_gb': row[4],
                    'hourly_price': float(row[5])
                }
                for row in cursor.fetchall()
            ]
    
    def get_latency_matrix(self) -> Dict[str, Dict[str, float]]:
        """지연시간 매트릭스 조회"""
        with self.conn.cursor() as cursor:
            cursor.execute("""
                SELECT source_region, target_region, avg_latency
                FROM cloud_latency
            """)
            
            matrix = {}
            for source, target, latency in cursor.fetchall():
                if source not in matrix:
                    matrix[source] = {}
                matrix[source][target] = float(latency)
            
            return matrix
    
    def close(self):
        self.conn.close()

class AggregatorOptimizer:
    """집계자 배치 최적화 클래스"""
    
    def __init__(self, request_data: Dict):
        self.db = DatabaseManager()
        self.federated_learning = request_data['federatedLearning']
        self.aggregator_config = request_data['aggregatorConfig']
        self.participants = self.federated_learning['participants']
        self.constraints = self._extract_constraints()
        
        self.price_data = self.db.get_cloud_prices()
        self.latency_matrix = self.db.get_latency_matrix()
        self.options = self._generate_options()
    
    def _extract_constraints(self) -> Dict:
        """aggregatorConfig에서 제약사항 추출 - 제약 완화"""
        config = self.aggregator_config
        return {
            # 제약사항을 매우 관대하게 설정하여 더 많은 옵션 포함
            'maxBudget': config.get('maxBudget', 100000000),  # 1억원으로 기본값 설정
            'maxLatency': config.get('maxLatency', 1000.0),   # 1000ms로 기본값 설정
            'minMemoryRequirement': config.get('minMemoryRequirement', 4),  # 4GB로 기본값 설정
        }
    
    def _generate_options(self) -> List[Dict]:
        """집계자 배치 옵션 생성 - 더 많은 옵션 포함"""
        options = []
        
        for price in self.price_data:
            region = price['region_name']
            
            # 지연시간 계산 - 데이터가 없어도 기본값 사용
            latencies = []
            for participant in self.participants:
                participant_region = participant.get('region', 'unknown')
                
                if (participant_region in self.latency_matrix and 
                    region in self.latency_matrix[participant_region]):
                    latencies.append(self.latency_matrix[participant_region][region])
            
            if not latencies:
                avg_latency = 5  # 기본값 1초
                max_latency = 5
            else:    
                avg_latency = sum(latencies) / len(latencies)
                max_latency = max(latencies)
                
            monthly_cost = price['hourly_price'] * 24 * 30 * USD_TO_KRW
            
            # 제약사항 확인 (매우 관대함)
            if (monthly_cost <= self.constraints['maxBudget'] and 
                avg_latency <= self.constraints['maxLatency'] and
                price['memory_gb'] >= self.constraints['minMemoryRequirement']):
                
                options.append({
                    'region': region,
                    'instanceType': price['instance_type'],
                    'cloudProvider': price['cloud_name'],
                    'cost': monthly_cost,
                    'avgLatency': avg_latency,
                    'maxLatency': max_latency,
                    'vcpu': price['v_cpu_count'],
                    'memory': price['memory_gb'],
                    'hourlyPrice': price['hourly_price']
                })
        
        print(f"디버그: 생성된 총 옵션 수: {len(options)}")
        return options
    
    def filter_options(self, options_list: List[Dict], constraints: Dict) -> List[Dict]:
        """사용자 제약조건에 맞는 옵션 필터링"""
        return [
            option for option in options_list
            if option['cost'] <= constraints['maxBudget'] and
               option['avgLatency'] <= constraints['maxLatency']
        ]

    
    # AggregatorOptimizer 클래스 내부에 아래 함수를 붙여넣거나 교체하세요.

    def nsga2_optimize(self) -> Tuple[List, List[Dict]]:
        """NSGA-II 최적화 실행 """
        if not self.options:
            raise ValueError("사용자 조건에 맞는 집계자 옵션이 없습니다.")
        
        filtered_options = self.options
        print(f"디버그: NSGA-II에 사용할 옵션 수: {len(filtered_options)}")
        
        # 옵션들의 분포 확인
        costs = [opt['cost'] for opt in filtered_options]
        latencies = [opt['avgLatency'] for opt in filtered_options]
        print(f"디버그: Cost 범위: {min(costs):.2f} ~ {max(costs):.2f}")
        print(f"디버그: Latency 범위: {min(latencies):.2f} ~ {max(latencies):.2f}")

        # DEAP 설정
        if hasattr(creator, "FitnessMulti"):
            del creator.FitnessMulti
        if hasattr(creator, "Individual"):
            del creator.Individual
            
        creator.create("FitnessMulti", base.Fitness, weights=(-1.0, -1.0))
        creator.create("Individual", list, fitness=creator.FitnessMulti) 

        toolbox = base.Toolbox()
        toolbox.register("attr_int", random.randint, 0, len(filtered_options) - 1)
        toolbox.register("individual", tools.initRepeat, creator.Individual, toolbox.attr_int, n=1)
        
        # 모집단 크기 조정
        population_size = min(200, max(50, len(filtered_options)))  # 최소 50, 최대 200
        toolbox.register("population", tools.initRepeat, list, toolbox.individual)

        # 정규화를 위한 최대값
        max_cost = max(costs) if costs else 1
        max_latency = max(latencies) if latencies else 1
        
        def evaluate_single_option(individual):
            option_index = individual[0]
            if option_index >= len(filtered_options):
                option_index = len(filtered_options) - 1
            option = filtered_options[option_index]
            # 정규화된 값 반환
            return (option['cost'] / max_cost, 
                    option['avgLatency'] / max_latency)

        toolbox.register("evaluate", evaluate_single_option)
        toolbox.register("mate", tools.cxUniform, indpb=0.5)
        toolbox.register("mutate", tools.mutUniformInt, low=0, up=len(filtered_options)-1, indpb=0.3)
        toolbox.register("select", tools.selNSGA2)
        
        # 초기 모집단 생성
        population = toolbox.population(n=population_size)
        
        # 초기 평가
        fitnesses = list(map(toolbox.evaluate, population))
        for ind, fit in zip(population, fitnesses):
            ind.fitness.values = fit
        
        # 진화 실행
        ngen = 200  # 300은 너무 많을 수 있음
        cxpb, mutpb = 0.7, 0.3
        
        for gen in range(ngen):
            offspring = toolbox.select(population, len(population))
            offspring = list(map(toolbox.clone, offspring))
            
            # 교차와 변이
            for child1, child2 in zip(offspring[::2], offspring[1::2]):
                if random.random() < cxpb:
                    toolbox.mate(child1, child2)
                    del child1.fitness.values
                    del child2.fitness.values
            
            for mutant in offspring:
                if random.random() < mutpb:
                    toolbox.mutate(mutant)
                    del mutant.fitness.values
            
            # 평가
            invalid_ind = [ind for ind in offspring if not ind.fitness.valid]
            fitnesses = map(toolbox.evaluate, invalid_ind)
            for ind, fit in zip(invalid_ind, fitnesses):
                ind.fitness.values = fit
            
            population[:] = offspring
            
            if gen % 20 == 0:
                fronts = tools.sortNondominated(population, len(population), first_front_only=True)
                print(f"디버그: Generation {gen}, First front size: {len(fronts[0])}")
        
        # 모든 Pareto Front 추출
        all_fronts = tools.sortNondominated(population, len(population), first_front_only=False)
        print(f"디버그: 총 front 수: {len(all_fronts)}")
        
        # 상위 3개 front에서 해 수집
        desired_solutions = 25  # 중복 제거 전에 더 많이 수집
        collected_individuals = []
        
        for i, front in enumerate(all_fronts[:3]):  # 상위 3개 front만
            print(f"디버그: Front {i} 크기: {len(front)}")
            collected_individuals.extend(front)
            if len(collected_individuals) >= desired_solutions:
                break
        
        # 인덱스 추출 및 중복 제거
        seen_options = set()
        final_indices = []
        
        for ind in collected_individuals[:desired_solutions]:
            index = ind[0]
            option = filtered_options[index]
            
            # 더 세밀한 중복 체크 (cloudProvider도 포함)
            option_key = (option['region'], option['instanceType'], option['cloudProvider'])
            
            if option_key not in seen_options:
                seen_options.add(option_key)
                final_indices.append(index)
                
                if len(final_indices) >= 20:  # 최종 목표 개수
                    break
        
        print(f"디버그: 수집된 개체 수: {len(collected_individuals)}")
        print(f"디버그: 중복 제거 후 최종 개수: {len(final_indices)}")
        
        # 만약 결과가 너무 적으면 다양성을 위해 추가
        if len(final_indices) < 20:
            print("디버그: 결과가 너무 적어 다양성 추가")
            # Cost 기준 상위 몇 개
            sorted_by_cost = sorted(range(len(filtered_options)), 
                                key=lambda i: filtered_options[i]['cost'])[:5]
            # Latency 기준 상위 몇 개
            sorted_by_latency = sorted(range(len(filtered_options)), 
                                    key=lambda i: filtered_options[i]['avgLatency'])[:5]
            
            for idx in sorted_by_cost + sorted_by_latency:
                if idx not in final_indices:
                    final_indices.append(idx)
                    if len(final_indices) >= 20:
                        break

        return final_indices, filtered_options
    
    

    def optimize(self) -> List[Dict]:
        try:
            # NSGA-II 최적화 실행
            # 반환값은 이제 '최적 인덱스 리스트'입니다. 변수명을 명확하게 변경합니다.
            optimal_indices, filtered_options = self.nsga2_optimize()
            
            # Pareto front의 모든 해를 결과로 변환
            pareto_solutions = []
            
            # 💥 더 명확하고 간결해진 로직
            # optimal_indices 리스트에서 인덱스를 하나씩 가져옵니다.
            for index in optimal_indices:
                # 해당 인덱스를 사용하여 바로 옵션 정보를 찾습니다.
                option = filtered_options[index]
                pareto_solutions.append(option)
            
            print(f"디버그: 최종 솔루션 수: {len(pareto_solutions)}")
            
            # 결과 포맷팅
            return self._format_results(pareto_solutions)
            
        except Exception as e:
            print(f"최적화 오류: {e}")
            # 오류 발생 시 상위 20개 옵션이라도 반환
            if self.options:
                print("오류 발생 - 기본 옵션으로 대체")
                sorted_options = sorted(self.options, key=lambda x: (x['cost'] * 0.4 + x['avgLatency'] * 0.6))
                # 오류 발생 시 반환 개수를 20개로 수정합니다.
                return self._format_results(sorted_options[:25])
            return []
    
    def _format_results(self, options: List[Dict]) -> List[Dict]:
        """결과를 프론트엔드 형식으로 포맷팅"""
        if not options:
            return []
        
        # 정규화를 위한 최대값 계산
        max_cost = max(opt['cost'] for opt in options) if options else 1
        max_latency = max(opt['avgLatency'] for opt in options) if options else 1
        
        # 가중합으로 점수 계산 및 정렬
        for opt in options:
            norm_cost = opt['cost'] / max_cost
            norm_latency = opt['avgLatency'] / max_latency
            opt['_score'] = 0.4 * norm_cost + 0.6 * norm_latency
        
        # 점수 기준으로 정렬 (낮을수록 좋음)
        options.sort(key=lambda x: x['_score'])
        
        # 최종 결과 형식
        results = []
        for i, opt in enumerate(options):
            results.append({
                'rank': i + 1,
                'region': opt['region'],
                'instanceType': opt['instanceType'],
                'cloudProvider': opt['cloudProvider'],
                'estimatedMonthlyCost': round(opt['cost'], 0),
                'estimatedHourlyPrice': round(opt['hourlyPrice'], 4),
                'avgLatency': round(opt['avgLatency'], 2),
                'maxLatency': round(opt['maxLatency'], 2),
                'vcpu': opt['vcpu'],
                'memory': opt['memory'],
                'recommendationScore': round((1 - opt['_score']) * 100, 1),  # 높을수록 좋음
            })
        
        return results
    
    def get_summary(self) -> Dict:
        """최적화 요약 정보"""
        return {
            'totalParticipants': len(self.participants),
            'participantRegions': list(set(p.get('region', 'unknown') for p in self.participants)),
            'totalCandidateOptions': len(self.price_data),
            'feasibleOptions': len(self.options),
            'constraints': self.constraints,
            'modelInfo': {
                'name': self.federated_learning.get('name', ''),
                'modelType': self.federated_learning.get('modelType', ''),
                'rounds': self.federated_learning.get('rounds', 0)
            },
            'optimizationMethod': 'NSGA-II (Non-dominated Sorting Genetic Algorithm II)'
        }
    
    def __del__(self):
        if hasattr(self, 'db'):
            self.db.close()

def main():
    """메인 함수 - API 요청 처리"""
    if len(sys.argv) != 3:
        print("사용법: python3 aggregator_optimization.py <input_file> <output_file>")
        sys.exit(1)
    
    try:
        # 백엔드에서 전달받은 요청 데이터 로드
        with open(sys.argv[1], 'r', encoding='utf-8') as f:
            request_data = json.load(f)
        
        # 최적화 실행
        optimizer = AggregatorOptimizer(request_data)
        optimization_results = optimizer.optimize()
        summary = optimizer.get_summary()
        
        # API 응답 형식으로 결과 구성
        response = {
            'status': 'success',
            'summary': summary,
            'optimizedOptions': optimization_results,
            'paretoFrontSize': len(optimization_results),
            'message': f'{len(optimization_results)}개의 최적 집계자 옵션을 찾았습니다.'
        }
        
        # 결과 저장
        with open(sys.argv[2], 'w', encoding='utf-8') as f:
            json.dump(response, f, indent=2, ensure_ascii=False)
        
        print(f"최적화 완료: {len(optimization_results)}개 옵션 생성")
        
    except Exception as e:
        # 에러 응답
        error_response = {
            'status': 'error',
            'message': str(e),
            'optimizedOptions': [],
            'paretoFrontSize': 0
        }
        
        with open(sys.argv[2], 'w', encoding='utf-8') as f:
            json.dump(error_response, f, indent=2, ensure_ascii=False)
        
        print(f"오류 발생: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()