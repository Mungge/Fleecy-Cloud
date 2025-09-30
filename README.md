<p align="center">
  <img src="https://github.com/user-attachments/assets/4a763cd4-2ed0-481b-aebc-1cd9fb60c5bc" style="width: 40%; max-width: 1000px;" alt="image" />
</p>

# Fleecy-Cloud

🌐 멀티 클라우드 인프라 기반 연합학습 환경 구축 플랫폼
> A Secure, Scalable, and Cloud-Native Federated Learning Platform Across Public and Private Clouds

## 👥 Team Members

| Profile | Role                                | Email | GitHub |
|:------:|:------------------------------------|:------|:--------|
| <p align="center"><img src="https://github.com/Jeon-Jinhyeok.png?size=80" width="80"/><br/><strong>전진혁</strong></p> | Team Leader / AI & System Architect | aqwstn@gmail.com | [@Jeon-Jinhyeok](https://github.com/Jeon-Jinhyeok) |
| <p align="center"><img src="https://github.com/kim-minkyoung.png?size=80" width="80"/><br/><strong>김민경</strong></p> | Backend Developer, CloudOps | decomin02@naver.com | [@Kim-Minkyoung](https://github.com/kim-minkyoung) |
| <p align="center"><img src="https://github.com/JAEIL1999.png?size=80" width="80"/><br/><strong>박재일</strong></p> | Backend Developer, MLOps | pkyj040410@gmail.com | [@JAEIL1999](https://github.com/JAEIL1999) |




---

## 1. 프로젝트 배경

### 1.1. 배경 소개

#### - 연합학습이란?
- **연합학습(Federated Learning)** 은 참여자(Client)와 집계자(Aggregator)로 구성된 분산학습 기술
- 기존의 중앙 집중형 기계학습(Machine Learning)과 달리 학습 데이터를 중앙 서버로 모으지 않고 각 참여자의 로컬 환경에서 모델을 학습한 후 **모델의 파라미터만 집계자에게 전송**하고, 원본 데이터는 로컬 환경에 유지
- 이러한 연합학습의 특성은 데이터 유출 위험을 감소시켜, 민감한 데이터를 다루는 의료, 제약, 스마트시티 등의 분야에서 활발하게 활용

#### - 멀티 클라우드란?
- **멀티 클라우드(Multi-Cloud)** 는 이형의 클라우드 플랫폼(퍼블릭 클라우드, 프라이빗 클라우드)을 통합하여 단일 클라우드 플랫폼처럼 활용할 수 있는 기술
- 이를 통해 **클라우드 벤더 종속성을 탈피**하고, **각 클라우드 플랫폼의 장점을 조합**할 수 있는 전략 수립 가능

#### - 멀티 클라우드 + 연합학습
- 멀티 클라우드와 연합학습을 결합한 플랫폼은 **개별 클라우드에서 획득할 수 있는 장점(비용 효율성, 접근성, 보안성 등)을 적용**하여 구축 가능
- ex) 개인의 민감한 의료 정보와 같이 강력한 보안성을 요구하는 경우 데이터를 프라이빗 클라우드에 저장하여 보안 강화, 다수의 데이터를 집계했을 때 가치를 발휘하는 질병 통계와 같은 정보는 퍼블릭 클라우드에 저장하여 접근성 및 활용도 높이는 전략 수립 

### 1.2. 국내외 시장 현황 및 문제점

#### (1) 비용 및 지연 최적화 부재
<p align="center">
<img src="https://github.com/user-attachments/assets/cefcc7c1-a191-45a8-b5c3-0345610de230" alt="비용 및 지연 최적화 부재" width="60%">
</p>

> 기존 클라우드 기반 연합학습의 경우, **단일 클라우드 플랫폼**에 한정되어 연합학습을 수행한다.  </br>이로인해 **리전별 비용 정책 차이와 네트워크 지연(latency)** 을 고려하지 못한다.
학습 참여자가 물리적으로 멀리 떨어져 있는 경우, 모델 파라미터  <br>전송에 지연이 발생하여 **학습 속도가 저하**된다. 또한 비용이 높은 리전에 집계자가 배포되면 **클라우드 비용이 불필요하게 증가**한다.

#### (2) 보안 취약성 및 데이터 프라이버시 문제
<p align="center">
  <img src="https://github.com/user-attachments/assets/a6d5c405-913e-48eb-8c44-c7f59e460df1" alt="보안 취약성" width="50%">
</p>

> 기존 클라우드 기반 연합학습의 경우, 단일 클라우드 플랫폼만 사용하여, 클라우드 플랫폼에 **학습 데이터의 업로드가 요구**된다.
</br>**학습 데이터가 클라우드 인프라 및 네트워크 환경에 노출**됨에 따라 세션 하이제킹, 중간자 공격(MITM), 무단 접근 등의 보안 문제가 발생할 수 있다.

#### (3) 동적 오케스트레이션 부재
<p align="center">
  <img src="https://github.com/user-attachments/assets/667f004b-9077-4a82-a394-823df0583052" alt="동적 오케스트레이션 부재" width="55%">
</p>

> 기존 연합학습 플랫폼의 경우 연합학습 태스크를**연합학습 참여자의 적절한 가상머신에 할당하는 동적 태스크 오케스트레이션이 부재**하다.</br>
컴퓨팅 리소스가 부족하거나, 상태가 Inactive인 가상머신을 선택하는 경우,**참여자 이탈(학습에 참여하지 못함) 및 연합학습 실패가 발생**할 수 있다.

### 1.3. 필요성과 기대효과

#### **필요성**
- **벤더 종속성 문제**: 특정 클라우드 서비스에 의존하면 리소스 활용과 확장성에 한계 발생
- **데이터 프라이버시**: 의료, 금융, 공공 데이터 등은 민감성이 높아 데이터 외부 반출이 금지되며, 데이터를 외부 인프라에 노출하지 않은 연합학습 환경 필요
- **운영 효율성**: 기존 단일 클라우드 기반 연합학습은 클라우드 비용· 학습시간 최적화가 부족해 비효율 발생
- **연합학습 안정성**: 기존 플랫폼은 VM/노드의 자원 상태나 장애를 실시간 반영하지 못해 학습 안정성이 저하됨
 
#### **기대효과**
- **벤더 종속성 해소**: 멀티 클라우드 네이티브 아키텍처를 통해 다양한 클라우드 리소스를 통합 활용함으로써 특정 벤더에 의존하지 않고 유연한 확장 가능  
- **데이터 프라이버시 강화**: 민감 데이터는 프라이빗 클라우드에 안전하게 보관하고, 글로벌 모델 집계만 퍼블릭 클라우드에서 수행하여 **데이터 외부 노출 위험 최소화**  
- **운영 효율성 향상**: 비용·지연을 고려한 최적의 집계자 배치 및 학습 자원 활용을 통해 **클라우드 비용 및 학습속도 최적화** 달성  
- **연합학습 안정성 확보**: 동적 오케스트레이션을 통해 VM/노드의 자원 상태와 장애를 실시간 반영하여 **안정적인 연합학습 수행** 보장  

---

## 2. 개발 목표

### 2.1. 목표 및 세부 내용

(1) 클라우드 별 비용 및 지연 시간을 고려한 멀티 클라우드 기반 연합학습 환경 구축

(2) 연합학습 집계자 - 연합학습 참여자 계층 기반의 멀티 클라우드 지원 연합학습 방법 도출

(3) 연합학습 참여자 모니터링을 통한 동적 태스크 오케스트레이션 기술 구현

### 2.2 기존 서비스 대비 차별성
**(1) 멀티 클라우드 환경 지원**

기존 IBM Federated Learning, AWS SageMaker, GCP Vertex AI는 단일 클라우드에 종속되어 벤더 락인 문제가 발생한다. 본 플랫폼은 AWS, GCP 등 이기종 퍼블릭 클라우드와 프라이빗 클라우드를 동시에 활용하여 벤더 종속성을 해결하고 각 클라우드의 장점을 선택적으로 활용할 수 있다.

**(2) 비용 및 지연 시간 최적화 지원**

기존 클라우드 기반 연합학습은 리전 간 비용 차이와 네트워크 지연을 고려하지 않는다. 본 플랫폼은 클라우드 운용 비용과 참여자 간 지연 시간을 종합 분석하여 최적의 집계자 배포 위치를 추천함으로써 비용 절감과 학습 속도 향상을 동시에 달성한다.

**(3) 계층 구조 기반 연합학습 수행**

기존 클라우드 기반 연합학습은 모든 프로세스를 단일 클라우드에서 처리하여 민감 데이터가 퍼블릭 클라우드에 노출된다. 본 플랫폼은 프라이빗 클라우드의 참여자에 학습 데이터를 격리하고, 퍼블릭 클라우드의 집계자는 모델 파라미터만 처리하도록 계층화하여 데이터 유출 위험을 최소화한다.

**(4) 동적 태스크 오케스트레이션 지원**

FedML 등 기존 프레임워크는 실시간 자원 모니터링 기반의 동적 작업 재배분이 제한적이다. 본 플랫폼은 참여자 가상머신의 CPU, GPU, 메모리 상태를 실시간 모니터링하고 자원 상태에 따라 학습 작업을 동적으로 할당하여 과부하를 방지하고 안정적인 연합학습 수행을 보장한다.


### 2.3. 사회적 가치 도입 계획

**(1) 공공성 강화**

본 플랫폼은 의료, 교육, 금융 등 민감한 개인정보를 다루는 분야에서 데이터 프라이버시를 보호하면서도 협업 학습을 가능하게 한다. 특히 의료 데이터와 같이 법적·윤리적 제약으로 인해 공유가 어려운 데이터를 활용한 AI 모델 개발을 지원함으로써, 질병 진단, 공공 보건, 교육 격차 해소 등 사회적 문제 해결에 기여한다.

**(2) 지속 가능성 확보**

멀티 클라우드 아키텍처를 통해 특정 클라우드 벤더에 종속되지 않는 장기적이고 확장 가능한 인프라를 제공한다. 이는 비용 최적화와 리소스 효율성을 높일 뿐만 아니라, 다양한 클라우드 플랫폼의 강점을 활용하여 지속 가능한 AI 생태계 구축을 가능하게 한다. 또한 동적 태스크 오케스트레이션을 통해 자원을 효율적으로 활용함으로써 환경적 지속 가능성에도 기여한다.

---

## 3. 시스템 설계

### 3.1. 유스케이스 다이어그램
<p align="center">
  <img width="60%" alt="유스케이스 다이어그램" src="https://github.com/user-attachments/assets/ca75fb9a-7299-439c-a76e-514773fffbcd" />
</p>

### 3.2. 시스템 구성도
<p align="center">
  <img src="https://github.com/user-attachments/assets/93b444a4-09e1-4028-991e-883de5102451" alt="시스템 아키텍처" width="70%">
</p>

본 프로젝트에서는 **퍼블릭 클라우드와 프라이빗 클라우드의 역할을 구분하여 연합학습을 수행할 수 있는 멀티 클라우드 기반 연합학습 시스템**을 구축한다.  

- 퍼블릭 클라우드에는 **연합학습 집계자(Aggregator)** 와 **글로벌 모델**이 배치되어, 모델 파라미터 집계와 글로벌 모델 업데이트를 담당한다.  
- 프라이빗 클라우드(OpenStack 기반)에는 **로컬 모델, 연합학습 참여자 VM, 학습 데이터셋**이 배치되어, 각 참여자는 로컬 환경에서 개별적으로 학습을 수행한다.  
- 학습된 파라미터는 퍼블릭 클라우드의 집계자에게 전송되며, 집계자는 이를 통합하여 글로벌 모델을 개선한다.  

이러한 계층 구조를 통해 데이터는 프라이빗 클라우드 내부에 안전하게 유지되면서도, 효율적인 연합학습을 지원한다.

#### 시스템 핵심 모듈
- **Aggregator Deployment Optimizer**  
  사용자 요구사항 기반 최적 집계자 명세(클라우드 리전, 스펙) 추천 및 배포를 통한 비용·학습시간 최적화  

- **Cloud Authenticator**  
  멀티 클라우드 연동을 위한 클라우드 인증 정보 관리 및 인증 정보 기반 연합학습 참여자 등록 지원  

- **Federated Learning Initializer**  
  연합학습 집계자 - 연합학습 집계자를 연계한 연합학습 수행을 위해 환경 설정(라이브러리 설치, 환경 변수, 학습 코드) 및 학습 수행 명령 전달  

- **Dynamic Task Orchestrator**  
  VM 자원 상태·장애 여부 기반 작업 할당, 실패 시 재시도로 안정성 보장  

- **Model Manager**  
  라운드별 글로벌 모델 관리, 저장·평가 지표 모니터링, 모델 다운로드 지원  

### 3.3. 사용 기술
| 분류 | 기술 |
| --- | --- |
| **Frontend** | Next.js 14, TypeScript, Tailwind CSS |
| **Backend** | Go (Gin) |
| **Cloud Platform** | AWS, GCP, OpenStack |
| **Infrastructure** | Terraform, Docker, Docker Compose |
| **Monitoring** | Prometheus, Grafana |
| **ML/AI Framework** | PyTorch, TensorFlow |
| **Federated Learning**|  Flower (Federated Learning) |
| **MLOps** | MLflow |
| **Database** | PostgreSQL |

---

## 4. 개발 결과

### 4.1. 전체 시스템 흐름도
<p align="center">
  <img width="90%" alt="image" src="https://github.com/user-attachments/assets/d165ae7e-2c41-4939-94a7-02c361da233d" />
</p>
</br>

### 4.2. 기능 설명

#### 1) 계층 구조 기반 연합학습 수행
<p align="center">
  <img width="50%" alt="계층 구조 기반 연합학습" src="https://github.com/user-attachments/assets/4f539384-170e-4b05-abda-1263db7c36ed" />
</p>

본 플랫폼은 **퍼블릭 클라우드와 프라이빗 클라우드의 역할을 명확히 분리**하여 연합학습을 수행한다.  

- **퍼블릭 클라우드**
  - **연합학습 집계자(Aggregator)**: 참여자 파라미터 수집 및 통합  
  - **글로벌 모델**: 라운드별 모델 저장 및 관리  
  - **특징**: 높은 접근성·확장성 제공  

- **프라이빗 클라우드(OpenStack 기반)**
  - **학습 데이터**: 민감한 학습 데이터 안전하게 보관  
  - **연합학습 참여자 VM**: 로컬 모델 학습 수행  
  - **특징**: 완전한 데이터 격리 보장  

> **프라이빗 클라우드**는 학습 데이터와 로컬 학습을 담당하고, **퍼블릭 클라우드**는 파라미터 집계 및 글로벌 모델 관리에 집중한다.  
이를 통해 데이터는 프라이빗 클라우드 내부에 안전하게 유지되면서도, 퍼블릭 클라우드의 확장성과 접근성을 활용할 수 있다.

#### 2) Cloud Authenticator
<p align="center">
  <img width="50%" height="893" alt="Cloud Authenticator 아키텍처" src="https://github.com/user-attachments/assets/4033841e-5ddb-4ac2-9517-0078ab210303" />
</p>

- **문제점**: 멀티 클라우드 환경에서는 AWS, Azure, GCP, OpenStack 등 각기 다른 클라우드 플랫폼별 인증 체계가 달라, 통합된 방식으로 연동·관리하기 어렵다.  
- **방법**: Cloud Authenticator를 통해 클라우드 자격 증명 파일을 업로드하고 API 기반으로 자동 검증 후, 본 플랫폼과 연동한다.  
  - 자격 증명 파일 업로드  
  - 클라우드 API 호출을 통한 자격 증명 검증
  - 검증 성공 시 해당 플랫폼과 연동 
- **효과**: 여러 클라우드 플랫폼을 **단일 인터페이스**에서 관리 가능하며, 멀티 클라우드 서비스를 활용할 수 있다.
  
#### 3) Aggregator Deployment Optimizer
<p align="center">
  <img width="50%" alt="집계자 배치 최적화 아키텍처" src="https://github.com/user-attachments/assets/73b6a64b-8bd7-443e-934a-787dec8c1657" />
</p>

- **방법**: **사용자 요구사항**(최대 허용 비용, 지연시간, 비용-지연시간 가중치 비율)을 반영하여 **집계자 최적 명세**(클라우드 리전, 스펙)**추천** 및 **배포 수행**  
- **핵심 기술**: **NSGA-II 다목적 최적화 알고리즘**을 적용하여 비용과 지연 시간을 동시에 최소화된 집계자 명세 추천, **Terraform**을 이용한 자동 배포
- **효과**: **클라우드 비용 절감 및 연합학습 학습속도 향상**
- **결과 예시**
<img width="40%" alt="배치 최적화 결과" src="https://github.com/user-attachments/assets/820e949e-e0ad-44e8-8a21-0214996f3254" />

#### 4) Dynamic Task Orchestrator
<p align="center">
  <img width="70%" alt="동적 태스크 오케스트레이션 플로우" src="https://github.com/user-attachments/assets/7efddea0-3555-47b2-af32-d507d26c95a6" />
</p>

- **방법**: 여러 대의 가상머신을 운영하는 참여자 VM의 자원 상태를 실시간 모니터링하고 조건에 따라 적절한 가상머신에 작업을 동적으로 할당  
  - **최소 사양 기반 필터링**: CPU, GPU, 메모리 등 기본 자원 요건 충족 여부 확인  
  - **원형 큐 기반 작업 할당**: 조건 충족 노드 중에서 균등하게 작업 분배
- **효과**: VM 장애나 자원 불균형 상황에서도 안정적인 연합학습을 수행할 수 있도록 지원
- **결과 예시** 
<img width="40%" alt="동적 태스크 오케스트레이션 결과" src="https://github.com/user-attachments/assets/3bb2d189-c447-446f-a6b8-cec63be2fe09" />

#### 5) Federated Learning Initializer
<p align="center">
  <img width="60%" alt="그림2" src="https://github.com/user-attachments/assets/efae453d-b781-4e6a-bf85-e903f95fa5d7" />
</p>
 
- **역할**: 연합학습 집계자와 연합학습 참여자를 연계하여 연합학습 환경 설정, 연합학습 수행, 연합학습 로그 모니터링
  - 학습할 모델, 집계자 주소, 하이퍼파라미터 등 필수 설정 자동화
  - **Flower Framework** 기반으로 연합학습을 수행하고, **SSH 연결**을 통해 환결 설정, 실행 지원 및 연합학습 로그 모니터링  
- **효과**: 사용자는 복잡한 환경 설정 과정 없이 손쉽게 연합학습을 실행할 수 있으며, 로그 모니터링을 통한 진행 상황 확인 가능
- **연합학습 모니터링 예시**
<img width="40%" alt="image" src="https://github.com/user-attachments/assets/2ea1d1da-e5dc-4b9a-9161-895e7a6db26f" />


#### 6) Model Manager
- **역할**: MLflow를 통한 연합학습 모델 관리 및 모니터링
  - 모델 성능 추적(각 학습 라운드마다 생성되는 글로벌 모델 성능 추적)
  - 사용자 요구사항에 따른 최적의 모델 추천 및 다운로드 기능 제공
- **효과**: 사용자는 **성능 추적 + 자동 추천 + 다운로드**를 통해 최적의 모델을 손쉽게 확보 가능
- **모델 모니터링 및 다운로드 예시**
<img width="40%" alt="image" src="https://github.com/user-attachments/assets/2baeccba-f078-4996-b037-4d7109f75da8" />


### 4.3. 사례 연구: COVID-19 진단 모델 구축
- **데이터셋**: [Kaggle Covid-19 Image Dataset](https://www.kaggle.com/datasets/pranavraikokte/covid19-image-dataset/data) 
- **시나리오**: 글로벌 의료기관 협력 (유럽, 아시아, 미국 리전에 분산 배치)
- **결과**: 데이터 프라이버시 완전 보장하면서 효율적인 글로벌 모델 학습 달성

### 실험 결과

#### 1) 비용 및 성능 최적화 평가

<img width="50%" alt="image" src="https://github.com/user-attachments/assets/d06a1a47-9524-4840-af36-3d1a89fc7399" />

| 시나리오 | 평균 학습 시간 | 월간 비용 | 개선율 |
| --- | --- | --- | --- |
| **최적화 전** | 52.17초 | 38,938원 | - |
| **비용 우선 최적화** | 53초 | 23,026원 | **최적화 전에 비해 비용 클라우드 비용 41% 절감** |
| **지연시간 우선 최적화** | 48.4초 | 31,450원 | **최적화 전에 비해 학습시간 7% 감소, 클라우드 비용 19% 절감** |

#### 2) 동적 태스크 오케스트레이션 평가
<img width="50%" alt="image" src="https://github.com/user-attachments/assets/a8dc592d-1bea-4ab3-b773-c63567b09f5e" />

| 방식 | 참여자 이탈률 | 연합학습 성공률 | 비고 |
| --- | --- | --- | --- |
| **무작위 선택** | 60% | 40% | **Inactive VM, 리소스 부족한 VM 선택 -> 참여자 이탈 및 연합학습 실패 발생** |
| **제안 알고리즘** | 0% | **100%** | **상태 및 자원 기반 작업 할당 및 재시도를 통해 연합학습 성공률 100% 달성** |

### 4.4 디렉터리 구조
```
.
├── Participant-Setting
├── README.md
├── asset
│   ├── cloud_price_AWS.csv
│   ├── cloud_price_GCP.csv
│   └── latency_results.csv
├── backend
│   ├── Dockerfile
│   ├── config
│   │   └── database.go
│   ├── env.example
│   ├── go.mod
│   ├── go.sum
│   ├── handlers
│   │   ├── aggregator
│   │   ├── auth
│   │   ├── clouds.go
│   │   ├── federated_learning_handler.go
│   │   ├── participant_handler.go
│   │   ├── ssh_keypair_handler.go
│   │   ├── templates
│   │   └── virtual_machine_handler.go
│   ├── initialization
│   │   └── app_init.go
│   ├── main.go
│   ├── middlewares
│   │   └── auth_middleware.go
│   ├── models
│   │   ├── aggregator.go
│   │   ├── cloud_connection.go
│   │   ├── cloud_latency.go
│   │   ├── cloud_price.go
│   │   ├── federated_learning.go
│   │   ├── participant.go
│   │   ├── participant_federated_learning.go
│   │   ├── provider.go
│   │   ├── refresh_token.go
│   │   ├── region.go
│   │   ├── ssh_keypair.go
│   │   └── user.go
│   ├── repository
│   │   ├── aggregator_repository.go
│   │   ├── cloud_latency.go
│   │   ├── cloud_price.go
│   │   ├── cloud_repository.go
│   │   ├── federated_learning_repository.go
│   │   ├── participant_repository.go
│   │   ├── provider_repository.go
│   │   ├── refresh_token_repository.go
│   │   ├── region_repository.go
│   │   ├── ssh_keypair_repository.go
│   │   └── user_repository.go
│   ├── routes
│   │   ├── aggregator_routes.go
│   │   ├── auth_routes.go
│   │   ├── cloud_routes.go
│   │   ├── federated_learning_routes.go
│   │   ├── participant_routes.go
│   │   ├── ssh_keypair_routes.go
│   │   └── virtual_machine_routes.go
│   ├── scripts
│   │   ├── aggregator_optimization.py
│   │   ├── requirements.txt
│   │   └── run_optimizer.sh
│   ├── services
│   │   ├── aggregator
│   │   ├── cloud_keypair_service.go
│   │   ├── data_initialization.go
│   │   ├── optimization.go
│   │   ├── prometheus_service.go
│   │   ├── ssh_keypair_service.go
│   │   ├── virtual_machine_selection.go
│   │   ├── virtual_machine_service.go
│   │   └── vm_types.go
│   ├── utils
│   │   ├── encryption.go
│   │   ├── jwt.go
│   │   ├── ssh_client.go
│   │   ├── ssh_keygen.go
│   │   ├── startup_script.go
│   │   └── terraform.go
│   └── validators
│       └── aggregator
├── docker-compose.yml
├── flower-demo
│   ├── flower_demo
│   │   ├── client_app.py
│   │   ├── server_app.py
│   │   └── task.py
│   └── pyproject.toml
├── frontend
│   ├── Dockerfile
│   ├── components.json
│   ├── eslint.config.mjs
│   ├── next.config.ts
│   ├── package-lock.json
│   ├── package.json
│   ├── postcss.config.mjs
│   ├── public
│   │   └── fleecy.ico
│   ├── src
│   │   ├── api
│   │   ├── app
│   │   ├── components
│   │   ├── contexts
│   │   ├── hooks
│   │   ├── lib
│   │   ├── middleware.ts
│   │   └── types
│   └── tsconfig.json
├── simulator
│   ├── Dockerfile
│   ├── aggregators
│   │   ├── __init__.py
│   │   ├── fedadagrad.py
│   │   ├── fedadam.py
│   │   ├── fedavg.py
│   │   ├── fedprox.py
│   │   └── fedyogi.py
│   ├── client.py
│   ├── docker-compose.yml
│   ├── model.py
│   ├── requirements.txt
│   └── server.py
├── structure.txt
└── terraform
    ├── aws
    │   ├── locals.tf
    │   ├── main.tf
    │   ├── outputs.tf
    │   ├── providers.tf
    │   ├── terraform.tfvars
    │   └── variables.tf
    ├── common
    │   └── scripts
    └── gcp
        ├── deploy.sh
        ├── main.tf
        ├── outputs.tf
        ├── providers.tf
        ├── terraform.tfvars
        └── variables.tf
```

---

## 5. 설치 및 실행 방법

### 사전 요구사항

- **클라우드 계정**: AWS, GCP
- **OpenStack 환경** (프라이빗 클라우드)
- **GitHub OAuth App** (인증용) (Optional)
- **PostgreSQL** 데이터베이스

### ⚙️ 환경 설정

#### 1) 백엔드 환경변수 (`backend/.env`)

```
DB_HOST=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_PORT=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
```

#### 2) 프론트엔드 환경변수 (`frontend/.env.local`)

```
NEXT_PUBLIC_API_URL=
```

### 실행

```bash
cd frontend
npm install
npm run dev:all
```
---

## 6. 소개 자료 및 시연 영상
### 6.1. 프로젝트 소개 자료
[📑뭉게구름 발표자료](docs/03.발표자료/발표자료.pdf)

### 6.2 시연 영상
[![🎥뭉게구름 졸업과제 영상](http://img.youtube.com/vi/KugrTo0gUVo/0.jpg)](https://youtu.be/KugrTo0gUVo)

---

## 7. 참고 문헌
[1]  B. McMahan, E. Moore, D. Ramage, S. Hampson, and B. A. y Arcas, "Communication-Efficient Learning of Deep Networks from Decentralized Data," Proceedings of the 20th International Conference on Artificial Intelligence and Statistics (AISTATS), PMLR 54, pp. 1273-1282, Apr. 2017.

[2] J. Dean, G. Corrado, R. Monga, K. Chen, M. Devin, Q. Le, M. Mao, M. Ranzato, A. Senior, P. Tucker, K. Yang, and A. Ng, "Large Scale Distributed Deep Networks," Proceedings of the 25th International Conference on Neural Information Processing Systems (NeurIPS), pp. 1223-1231, Dec. 2012.

[3] L. Yuan, et al., "Decentralized federated learning: A survey and perspective," IEEE Internet of Things Journal, vol. 11, no. 21, pp. 34617-34638, 2024.

[4] J. Hong, T. Dreibholz, J. A. Schenkel, and J. A. Hu, “An Overview of Multi-cloud Computing,” Web, Artificial Intelligence and Network Applications, pp. 1055-1068, 2019.

[5] J. Alonso, et al., "Understanding the challenges and novel architectural models of multi-cloud native applications – a systematic literature review," Journal of Cloud Computing, vol. 12, no. 1, p. 6, 2023.

[6] AWS, “클라우드 컴퓨팅 서비스 - Amazon Web Services(AWS),” [Online]. Available: https://aws.amazon.com. [Accessed: Sep. 11, 2025].

[7] Microsoft, “Azure란? | Microsoft Azure,” [Online]. Available: https://azure.microsoft.com/ko-kr/resources/cloud-computing-dictionary/what-is-azure. [Accessed: Sep. 11, 2025].

[8] Google, “클라우드 컴퓨팅 서비스 | Google Cloud,” [Online]. Available: https://cloud.google.com.  [Accessed: Sep. 11, 2025].

[9] IBM, “Welcome to IBM Federated Learning — ibm-federated-learning,” [Online]. Available: https://ibmfl-api-docs.res.ibm.com. [Accessed: Sep. 11, 2025].

[10] AWS, “Amazon SageMaker AI Documentation,” [Online]. Available: https://docs.aws.amazon.com/sagemaker. [Accessed: Sep. 11, 2025].

[11] Google, “Vertex AI 문서 | Google Cloud,” [Online]. Available: https://cloud.google.com/vertex-ai/docs. [Accessed: Sep. 11, 2025].

[12] FedML, “TensorOpera® Documentation,” [Online]. Available: https://doc.fedml.ai/ [Accessed: Sep. 11, 2025].

[13] J. Proudman, “Openstack Docs: 2025.1,” [Online]. Available: https://docs.openstack.org. [Accessed: Sep. 11, 2025].

[14] Prometheus Authors, “Overview | Prometheus,” [Online]. Available: https://prometheus.io/docs. [Accessed: Sep. 13, 2025].

[15] Grafana, “Technical documentation | Grafana Labs,” [Online]. Available: https://grafana.com/docs. [Accessed: Sep. 11, 2025].

[16] Hashicorp, “Terraform | Terraform | HashiCorp Developer,” [Online]. Available: https://developer.hashicorp.com/terraform. [Accessed: Sep. 11, 2025].

[17] Flower, “Flower Documentation,” [Online]. Available: https://flower.dev/docs. [Accessed: Sep. 11, 2025].

[18] MLflow, “MLflow,” [Online]. Available: https://mlflow.org. [Accessed: Sep. 11, 2025].

[19] PyTorch, “PyTorch documentation - PyTorch 2.8 documentation,” [Online]. Available: https://pytorch.org/docs. [Accessed: Sep. 11, 2025].

[20] TensorFlow, “TensorFlow,” [Online]. Available: https://www.tensorflow.org. [Accessed: Sep. 11, 2025].

[21] Docker, “Docker Docs,” [Online]. Available: https://docs.docker.com. [Accessed: Sep. 11, 2025].

[22] J.Liu, and X.Chen, “An Improved NSGA-II Algorithm Based on Crowding Distance Elimination Strategy,” International Journal of Computational Intelligence Systems, Vol.12, No.2, pp.513-518, 2019.

[23] Z.Osika, P.Koch, and T.Wagner, “What lies beyond the Pareto front? A survey on decision-support methods for multi-objective optimization,” in arXiv preprint, pp.1-9, 2023.

[24] Kaggle, “Covid-19 Image Dataset,” [Online]. Available: https://www.kaggle.com/datasets/pranavraikokte/covid19-image-dataset/data. [Accessed: Sep. 11, 2025].

[25] N. Gavric, A. Shalaginov, A. Andrushevich, A. Rumsch, and A. Paice, "Enhancing International Data Spaces Security: A STRIDE Framework Approach," Preprints, 2024.

[26] ALI-POUR, Amir, et al. Towards a distributed federated learning aggregation placement using particle swarm intelligence. arXiv preprint arXiv:2504.16227, 2025.
