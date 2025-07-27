// components/dashboard/federated-learning/constants.ts

// 집계 알고리즘 목록
export const AGGREGATION_ALGORITHMS = [
    { id: "fedavg", name: "FedAvg" },
    { id: "fedprox", name: "FedProx" },
    { id: "scaffold", name: "SCAFFOLD" },
    { id: "fedopt", name: "FedOpt" },
  ];
  
  // 지원하는 모델 유형
  export const MODEL_TYPES = [
    { id: "image_classification", name: "이미지 분류" },
    { id: "nlp", name: "자연어 처리" },
    { id: "tabular", name: "테이블 형식 데이터" },
  ];
  
  // 파일 형식
  export const SUPPORTED_FILE_FORMATS = ".h5,.pb,.pt,.pth,.onnx,.pkl";
  
  // 폼 기본값
  export const FORM_DEFAULTS = {
    ALGORITHM: "fedavg",
    ROUNDS: 10,
  } as const;