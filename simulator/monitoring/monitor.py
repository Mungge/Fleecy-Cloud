import os
import time
import threading
import csv
from pathlib import Path
import psutil

class CPUMonitor:
    """컨테이너 CPU 사용량 모니터링 클래스 (psutil 사용)"""
    
    def __init__(self, container_name, output_file, interval=1.0):
        self.container_name = container_name
        self.output_file = output_file
        self.interval = interval
        self.is_running = False
        self.thread = None
        
        # 출력 디렉토리 생성
        Path(os.path.dirname(output_file)).mkdir(parents=True, exist_ok=True)
        
        # CSV 헤더 생성
        with open(self.output_file, 'w', newline='') as f:
            writer = csv.writer(f)
            writer.writerow(['timestamp', 'cpu_percent', 'aggregator_type', 'container_name'])
    
    def _monitor(self):
        """CPU 사용량을 주기적으로 측정하고 CSV 파일에 기록"""
        aggregator_type = os.environ.get("AGGREGATOR_TYPE", "unknown")
        
        while self.is_running:
            try:
                # CPU 사용량 계산 (psutil 사용)
                cpu_percent = psutil.cpu_percent(interval=0.1)
                
                # 현재 시간 기록
                timestamp = time.time()
                
                # CSV에 기록
                with open(self.output_file, 'a', newline='') as f:
                    writer = csv.writer(f)
                    writer.writerow([timestamp, cpu_percent, aggregator_type, self.container_name])
                
            except Exception as e:
                print(f"Error monitoring CPU: {e}")
            
            # 측정 간격만큼 대기
            time.sleep(self.interval)
    
    def start(self):
        """모니터링 시작"""
        if not self.is_running:
            self.is_running = True
            self.thread = threading.Thread(target=self._monitor)
            self.thread.daemon = True
            self.thread.start()
            print(f"Started CPU monitoring for {self.container_name}")
    
    def stop(self):
        """모니터링 중지"""
        if self.is_running:
            self.is_running = False
            if self.thread:
                self.thread.join(timeout=2.0)
            print(f"Stopped CPU monitoring for {self.container_name}")

class AggregatorMonitor:
    """다양한 Aggregator 방식의 CPU 사용량을 비교하는 클래스"""
    
    def __init__(self, results_dir="./results"):
        self.results_dir = results_dir
        Path(results_dir).mkdir(parents=True, exist_ok=True)
    
    def analyze_results(self):
        """수집된 CPU 사용량 데이터 분석"""
        import pandas as pd
        import matplotlib.pyplot as plt
        
        # 서버 CPU 사용량 데이터 파일 리스트
        cpu_files = list(Path(self.results_dir).glob("server_cpu_*.csv"))
        
        if not cpu_files:
            print("No CPU usage data files found")
            return
        
        # 모든 파일의 데이터 읽기
        all_data = []
        for file in cpu_files:
            try:
                df = pd.read_csv(file)
                # 파일 이름에서 Aggregator 타입 추출
                agg_type = file.stem.split('_')[-1]
                if 'aggregator_type' not in df.columns:
                    df['aggregator_type'] = agg_type
                all_data.append(df)
            except Exception as e:
                print(f"Error reading file {file}: {e}")
        
        if not all_data:
            print("No data could be loaded from CSV files")
            return
        
        # 모든 데이터 결합
        combined_data = pd.concat(all_data, ignore_index=True)
        
        # Aggregator 타입별 평균 CPU 사용량 계산
        agg_stats = combined_data.groupby('aggregator_type')['cpu_percent'].agg(['mean', 'min', 'max', 'std']).reset_index()
        
        # 결과 출력
        print("\nAggregator CPU Usage Statistics:")
        print(agg_stats)
        
        # 결과 저장
        agg_stats.to_csv(f"{self.results_dir}/aggregator_cpu_stats.csv", index=False)
        
        # 시각화
        plt.figure(figsize=(12, 6))
        
        # 시간에 따른 CPU 사용량 그래프
        plt.subplot(1, 2, 1)
        for agg_type, group in combined_data.groupby('aggregator_type'):
            plt.plot(group['timestamp'] - group['timestamp'].min(), group['cpu_percent'], label=agg_type)
        plt.xlabel('Time (seconds)')
        plt.ylabel('CPU Usage (%)')
        plt.title('CPU Usage Over Time by Aggregator Type')
        plt.legend()
        
        # 평균 CPU 사용량 바 차트
        plt.subplot(1, 2, 2)
        plt.bar(agg_stats['aggregator_type'], agg_stats['mean'], yerr=agg_stats['std'], capsize=5)
        plt.xlabel('Aggregator Type')
        plt.ylabel('Average CPU Usage (%)')
        plt.title('Average CPU Usage by Aggregator Type')
        
        plt.tight_layout()
        plt.savefig(f"{self.results_dir}/aggregator_cpu_comparison.png")
        plt.close()
        
        print(f"Results saved to {self.results_dir}")

def main():
    """모니터링 메인 함수 - 독립 실행 시 사용"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Analyze Aggregator CPU Usage')
    parser.add_argument('--results-dir', type=str, default='./results', help='Results directory')
    args = parser.parse_args()
    
    monitor = AggregatorMonitor(results_dir=args.results_dir)
    monitor.analyze_results()

if __name__ == "__main__":
    main()