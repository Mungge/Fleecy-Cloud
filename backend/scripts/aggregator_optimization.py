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


def log_print(message):
        print(message)  # 터미널에도 출력
        with open(log_file_path, 'a', encoding='utf-8') as log_f:
            log_f.write(f"{message}\n")
            log_f.flush()  # 즉시 파일에 쓰기
            
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
                SELECT cloud_name, region_name, instance_type, vcpu_count, 
                       memory_gb, on_demand_price
                FROM cloud_price
                ORDER BY cloud_name, region_name, on_demand_price
            """)
            return [
                {
                    'cloud_name': row[0],
                    'region_name': row[1], 
                    'instance_type': row[2],
                    'vcpu_count': row[3],
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
        """aggregatorConfig에서 제약사항 추출"""
        config = self.aggregator_config
        return {
            'maxBudget': config.get('maxBudget', float('inf')),
            'maxLatency': config.get('maxLatency', float('inf')),
        }
    
    def _generate_options(self) -> List[Dict]:
        """집계자 배치 옵션 생성"""
        options = []
        
        for price in self.price_data:
            region = price['region_name']
            
            # 평균 지연시간 계산
            latencies = []
            for participant in self.participants:
                participant_region = participant.get('openstack_region', 'unknown') # 수정 필요
                if (participant_region in self.latency_matrix and 
                    region in self.latency_matrix[participant_region]):
                    latencies.append(self.latency_matrix[participant_region][region])
            
            if not latencies:
                continue  # 지연시간 데이터가 없으면 스킵
                
            avg_latency = sum(latencies) / len(latencies)
            max_latency = max(latencies)
            monthly_cost = price['hourly_price'] * 24 * 30 * USD_TO_KRW
            
            # 제약사항 확인
            if (monthly_cost <= self.constraints['maxBudget'] and 
                avg_latency <= self.constraints['maxLatency']):
                
                options.append({
                    'region': region,
                    'instanceType': price['instance_type'],
                    'cloudProvider': price['cloud_name'],
                    'cost': monthly_cost,
                    'avgLatency': avg_latency,
                    'maxLatency': max_latency,
                    'vcpu': price['vcpu_count'],
                    'memory': price['memory_gb'],
                    'hourlyPrice': price['hourly_price']
                })
        
        return options
    
    def optimize(self) -> List[Dict]:
        """NSGA-II 최적화 실행하여 옵션 리스트 반환"""
        if not self.options:
            return []
        
        # 옵션이 적으면 그대로 반환
        if len(self.options) <= 20:
            log_print(f"🔍 [MAIN] 최적화 실행 시작")
            return self._format_results(self.options)
        
        # NSGA-II 설정
        creator.create("FitnessMulti", base.Fitness, weights=(-1.0, -1.0))
        creator.create("Individual", list, fitness=creator.FitnessMulti)
        
        toolbox = base.Toolbox()
        toolbox.register("attr_int", random.randint, 0, len(self.options) - 1)
        toolbox.register("individual", tools.initRepeat, creator.Individual, toolbox.attr_int, 1)
        toolbox.register("population", tools.initRepeat, list, toolbox.individual, 50)
        toolbox.register("evaluate", self._evaluate)
        toolbox.register("mate", tools.cxUniform, indpb=0.5)
        toolbox.register("mutate", tools.mutUniformInt, low=0, up=len(self.options)-1, indpb=0.5)
        toolbox.register("select", tools.selNSGA2)
        
        # 진화 알고리즘 실행
        population = toolbox.population()
        algorithms.eaMuPlusLambda(population, toolbox, mu=50, lambda_=75, 
                                cxpb=0.8, mutpb=0.1, ngen=100, verbose=False)
        
        # 파레토 최적해 추출
        pareto_front = tools.sortNondominated(population, len(population), first_front_only=True)[0]
        
        # 중복 제거
        unique_options = []
        seen = set()
        for individual in pareto_front:
            option = self.options[individual[0]]
            key = (option['region'], option['instanceType'])
            if key not in seen:
                seen.add(key)
                unique_options.append(option)
        
        return self._format_results(unique_options)
    
    def _format_results(self, options: List[Dict]) -> List[Dict]:
        """결과를 프론트엔드 형식으로 포맷팅"""
        if not options:
            return []
        
        # 가중합으로 정렬 (비용 40%, 평균지연시간 60%)
        max_cost = max(opt['cost'] for opt in options)
        max_latency = max(opt['avgLatency'] for opt in options)
        
        for opt in options:
            norm_cost = opt['cost'] / max_cost
            norm_latency = opt['avgLatency'] / max_latency
            opt['_score'] = 0.4 * norm_cost + 0.6 * norm_latency
        
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
                'recommendationScore': round((1 - opt['_score']) * 100, 1)  # 높을수록 좋음
            })
        
        return results
    
    def _evaluate(self, individual) -> Tuple[float, float]:
        """개체 평가 (비용, 평균지연시간)"""
        option = self.options[individual[0]]
        return option['cost'], option['avgLatency']
    
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
            }
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
            'message': f'{len(optimization_results)}개의 최적화된 집계자 옵션을 찾았습니다.'
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
            'optimizedOptions': []
        }
        
        with open(sys.argv[2], 'w', encoding='utf-8') as f:
            json.dump(error_response, f, indent=2, ensure_ascii=False)
        
        print(f"오류 발생: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()