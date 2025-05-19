import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
import os

# 영어 폰트로 변경
plt.rcParams['font.family'] = 'DejaVu Sans'

# 현재 디렉토리 출력
print(f"Current working directory: {os.getcwd()}")
print(f"Files in directory: {os.listdir('.')}")

# CSV 파일 읽기
csv_file = './results/server_cpu_fedprox.csv'
if os.path.exists(csv_file):
    df = pd.read_csv(csv_file)
    print(f"CSV file loaded successfully. Shape: {df.shape}")
    
    # 데이터 확인
    print("\nFirst 5 rows:")
    print(df.head())
    
    print("\nData types:")
    print(df.dtypes)
    
    print("\nCPU usage stats:")
    print(df['cpu_percent'].describe() if 'cpu_percent' in df.columns else "No 'cpu_percent' column found.")
    
    # 시각화
    plt.figure(figsize=(12, 10))
    
    # 시간 컬럼 설정
    if 'timestamp' in df.columns:
        df['relative_time'] = df['timestamp'] - df['timestamp'].min()
        time_column = 'relative_time'
    else:
        time_column = df.index
        df['relative_time'] = range(len(df))
    
    cpu_col = 'cpu_percent' if 'cpu_percent' in df.columns else df.columns[1]
    
    # 1. Line plot
    plt.subplot(2, 2, 1)
    plt.plot(df['relative_time'], df[cpu_col], 'b-', linewidth=1)
    plt.title('Fedprox Server CPU Usage')
    plt.xlabel('Time (Relative)')
    plt.ylabel('CPU Usage (%)')
    plt.grid(True)
    
    # 2. Moving average
    plt.subplot(2, 2, 2)
    window_size = max(1, len(df) // 20)
    df['cpu_smooth'] = df[cpu_col].rolling(window=window_size).mean()
    plt.plot(df['relative_time'], df[cpu_col], 'b-', alpha=0.3, linewidth=1, label='Raw')
    plt.plot(df['relative_time'], df['cpu_smooth'], 'r-', linewidth=2, label='Moving Avg')
    plt.title('CPU Usage with Moving Average')
    plt.xlabel('Time (Relative)')
    plt.ylabel('CPU Usage (%)')
    plt.legend()
    plt.grid(True)
    
    # 3. Histogram
    plt.subplot(2, 2, 3)
    sns.histplot(df[cpu_col].dropna(), bins=30, kde=True)
    plt.title('CPU Usage Distribution')
    plt.xlabel('CPU Usage (%)')
    plt.ylabel('Frequency')
    
    # 4. Boxplot
    plt.subplot(2, 2, 4)
    sns.boxplot(y=df[cpu_col].dropna())
    plt.title('CPU Usage Boxplot')
    plt.ylabel('CPU Usage (%)')
    
    # 저장 및 표시
    plt.tight_layout()
    plt.savefig('fedprox_cpu_analysis.png', dpi=300)
    print("\nGraph saved as 'fedprox_cpu_analysis.png'.")
    
    # 추가 통계
    print("\nExtra statistics:")
    print(f"Mean: {df[cpu_col].mean():.2f}%")
    print(f"Max: {df[cpu_col].max():.2f}%")
    print(f"Min: {df[cpu_col].min():.2f}%")
    print(f"Std Dev: {df[cpu_col].std():.2f}%")
else:
    print(f"File '{csv_file}' not found. Check the path.")
