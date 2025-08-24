// utils/modelDefinitionParser.ts

export interface ModelAnalysis {
    totalParams: number;
    modelSizeBytes: number;
    layerInfo: {
      name: string;
      type: string;
      params: number;
    }[];
    framework: 'pytorch' | 'tensorflow' | 'unknown';
  }
  
  /**
   * 모델 정의 파일(.py)을 분석하여 대략적인 모델 크기를 추정합니다.
   * 간단한 정규식 기반 파싱으로 주요 레이어들의 파라미터 수를 계산합니다.
   */
  export const analyzeModelDefinition = async (file: File): Promise<ModelAnalysis> => {
    const content = await file.text();
    
    const analysis: ModelAnalysis = {
      totalParams: 0,
      modelSizeBytes: 0,
      layerInfo: [],
      framework: detectFramework(content)
    };
  
    if (analysis.framework === 'pytorch') {
      analyzePyTorchModel(content, analysis);
    } else if (analysis.framework === 'tensorflow') {
      analyzeTensorFlowModel(content, analysis);
    } else {
      // 알 수 없는 프레임워크인 경우 기본값 사용
      analysis.totalParams = 1000000; // 1M 파라미터 기본값
    }
  
    // 파라미터가 너무 적게 계산된 경우 (transfer learning 등) 최소값 보장
    if (analysis.totalParams < 100000) {
      const minParams = estimateMinimumParams(content);
      analysis.totalParams = Math.max(analysis.totalParams, minParams);
      
      if (minParams > analysis.totalParams) {
        analysis.layerInfo.push({
          name: 'Estimated additional params',
          type: 'Estimated',
          params: minParams - analysis.totalParams
        });
      }
    }
  
    // 파라미터 수에서 바이트 크기 계산 (FP32 기준: 4 bytes per parameter)
    analysis.modelSizeBytes = analysis.totalParams * 4;
    
    return analysis;
  };
  
  /**
   * 모델의 최소 파라미터 수 추정 (변수명, 주석 등 기반)
   */
  const estimateMinimumParams = (content: string): number => {
    // num_classes 추출
    const numClassesMatch = content.match(/num_classes\s*=\s*(\d+)/);
    const numClasses = numClassesMatch ? parseInt(numClassesMatch[1]) : 3;
    
    // 이미지 크기 관련 정보 추출
    const hasImageProcessing = content.includes('224') || content.includes('ImageNet') || 
                              content.includes('Conv2d') || content.includes('AdaptiveAvgPool');
    
    // Transfer learning 감지
    const isTransferLearning = content.includes('pretrained') || content.includes('weights=') ||
                              content.includes('requires_grad = False') || content.includes('.features');
    
    if (isTransferLearning && hasImageProcessing) {
      // Transfer learning + 이미지: 최소 1M 파라미터
      return 1000000;
    } else if (hasImageProcessing) {
      // 이미지 모델: 최소 500K 파라미터
      return 500000;
    } else {
      // 일반 모델: 클래스 수 기반 추정
      return Math.max(numClasses * 1000, 100000);
    }
  };
  
  /**
   * 프레임워크 감지
   */
  const detectFramework = (content: string): 'pytorch' | 'tensorflow' | 'unknown' => {
    if (content.includes('import torch') || content.includes('nn.Module')) {
      return 'pytorch';
    }
    if (content.includes('tensorflow') || content.includes('keras')) {
      return 'tensorflow';
    }
    return 'unknown';
  };
  
  /**
   * Pretrained 모델 감지 및 파라미터 수 추정
   */
  const detectPretrainedModels = (content: string) => {
    const pretrainedModels: { name: string; params: number }[] = []; // 타입 정의 추가
    
    // VGG 모델들
    if (content.includes('models.vgg16') || content.includes('vgg16(')) {
        pretrainedModels.push({
            name: 'VGG16',
            params: 138357544 // 약 138M 파라미터
        });
    }
    if (content.includes('models.vgg19') || content.includes('vgg19(')) {
        pretrainedModels.push({
            name: 'VGG19',
            params: 143667240 // 약 144M 파라미터
        });
    }
  
    // ResNet 모델들
    if (content.includes('models.resnet18') || content.includes('resnet18(')) {
      pretrainedModels.push({
        name: 'ResNet18',
        params: 11689512 // 약 11.7M 파라미터
      });
    }
    if (content.includes('models.resnet50') || content.includes('resnet50(')) {
      pretrainedModels.push({
        name: 'ResNet50',
        params: 25557032 // 약 25.6M 파라미터
      });
    }
  
    // EfficientNet 모델들
    if (content.includes('models.efficientnet_b0') || content.includes('efficientnet_b0(')) {
      pretrainedModels.push({
        name: 'EfficientNet-B0',
        params: 5288548 // 약 5.3M 파라미터
      });
    }
  
    // BERT 계열 (transformers 라이브러리)
    if (content.includes('AutoModel') || content.includes('BertModel')) {
      pretrainedModels.push({
        name: 'BERT-base',
        params: 110000000 // 약 110M 파라미터 (기본값)
      });
    }
  
    // GPT 계열
    if (content.includes('GPT2') || content.includes('gpt2')) {
      pretrainedModels.push({
        name: 'GPT-2',
        params: 117000000 // 약 117M 파라미터 (base 모델)
      });
    }
  
    return pretrainedModels;
  };
  
  /**
   * PyTorch 모델 분석
   */
  const analyzePyTorchModel = (content: string, analysis: ModelAnalysis) => {
    // Pretrained 모델 사용 여부 확인
    const pretrainedModels = detectPretrainedModels(content);
    pretrainedModels.forEach(model => {
      analysis.layerInfo.push({
        name: `Pretrained ${model.name}`,
        type: 'Pretrained',
        params: model.params
      });
      analysis.totalParams += model.params;
    });
  
    // Linear/Dense 레이어 찾기 - nn.Linear(in_features, out_features)
    const linearMatches = content.match(/nn\.Linear\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)/g);
    if (linearMatches) {
      linearMatches.forEach(match => {
        const params = match.match(/(\d+)/g);
        if (params && params.length >= 2) {
          const inFeatures = parseInt(params[0]);
          const outFeatures = parseInt(params[1]);
          const paramCount = inFeatures * outFeatures + outFeatures; // weights + bias
          
          analysis.layerInfo.push({
            name: `Linear(${inFeatures}, ${outFeatures})`,
            type: 'Linear',
            params: paramCount
          });
          analysis.totalParams += paramCount;
        }
      });
    }
  
    // Sequential 내부의 Linear 레이어들도 찾기
    const sequentialMatches = content.match(/nn\.Sequential\s*\(([\s\S]*?)\)/g);
    if (sequentialMatches) {
      sequentialMatches.forEach(seqMatch => {
        const innerLinearMatches = seqMatch.match(/nn\.Linear\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)/g);
        if (innerLinearMatches) {
          innerLinearMatches.forEach(match => {
            const params = match.match(/(\d+)/g);
            if (params && params.length >= 2) {
              const inFeatures = parseInt(params[0]);
              const outFeatures = parseInt(params[1]);
              const paramCount = inFeatures * outFeatures + outFeatures;
              
              analysis.layerInfo.push({
                name: `Linear(${inFeatures}, ${outFeatures})`,
                type: 'Linear',
                params: paramCount
              });
              analysis.totalParams += paramCount;
            }
          });
        }
      });
    }
  
    // Conv2d 레이어 찾기 - nn.Conv2d(in_channels, out_channels, kernel_size)
    const convMatches = content.match(/nn\.Conv2d\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+|\(\d+,\s*\d+\))/g);
    if (convMatches) {
      convMatches.forEach(match => {
        const params = match.match(/(\d+)/g);
        if (params && params.length >= 3) {
          const inChannels = parseInt(params[0]);
          const outChannels = parseInt(params[1]);
          const kernelSize = parseInt(params[2]);
          const paramCount = inChannels * outChannels * kernelSize * kernelSize + outChannels;
          
          analysis.layerInfo.push({
            name: `Conv2d(${inChannels}, ${outChannels}, ${kernelSize})`,
            type: 'Conv2d',
            params: paramCount
          });
          analysis.totalParams += paramCount;
        }
      });
    }
  
    // Embedding 레이어 찾기
    const embeddingMatches = content.match(/nn\.Embedding\s*\(\s*(\d+)\s*,\s*(\d+)\s*\)/g);
    if (embeddingMatches) {
      embeddingMatches.forEach(match => {
        const params = match.match(/(\d+)/g);
        if (params && params.length >= 2) {
          const vocabSize = parseInt(params[0]);
          const embeddingDim = parseInt(params[1]);
          const paramCount = vocabSize * embeddingDim;
          
          analysis.layerInfo.push({
            name: `Embedding(${vocabSize}, ${embeddingDim})`,
            type: 'Embedding',
            params: paramCount
          });
          analysis.totalParams += paramCount;
        }
      });
    }
  };
  
  /**
   * TensorFlow/Keras 모델 분석
   */
  const analyzeTensorFlowModel = (content: string, analysis: ModelAnalysis) => {
    // Dense 레이어 찾기
    const denseMatches = content.match(/Dense\s*\(\s*(\d+)/g);
    if (denseMatches) {
      // 간단한 추정: 각 Dense 레이어가 이전 레이어와 연결되어 있다고 가정
      let prevSize = 128; // 기본값
      denseMatches.forEach(match => {
        const sizeMatch = match.match(/(\d+)/);
        if (sizeMatch) {
          const currentSize = parseInt(sizeMatch[1]);
          const paramCount = prevSize * currentSize + currentSize;
          
          analysis.layerInfo.push({
            name: `Dense(${currentSize})`,
            type: 'Dense',
            params: paramCount
          });
          analysis.totalParams += paramCount;
          prevSize = currentSize;
        }
      });
    }
  
    // Conv2D 레이어 찾기 (간단한 추정)
    const conv2dMatches = content.match(/Conv2D\s*\(\s*(\d+)/g);
    if (conv2dMatches) {
      conv2dMatches.forEach(match => {
        const sizeMatch = match.match(/(\d+)/);
        if (sizeMatch) {
          const filters = parseInt(sizeMatch[1]);
          const paramCount = filters * 3 * 3 * 3 + filters; // 3x3 kernel, 3 input channels 가정
          
          analysis.layerInfo.push({
            name: `Conv2D(${filters})`,
            type: 'Conv2D',
            params: paramCount
          });
          analysis.totalParams += paramCount;
        }
      });
    }
  };
  
  /**
   * 모델 크기를 사람이 읽기 쉬운 형태로 포맷
   */
  export const formatModelSize = (params: number): string => {
    if (params < 1000) {
      return `${params} 파라미터`;
    } else if (params < 1000000) {
      return `${(params / 1000).toFixed(1)}K 파라미터`;
    } else if (params < 1000000000) {
      return `${(params / 1000000).toFixed(1)}M 파라미터`;
    } else {
      return `${(params / 1000000000).toFixed(1)}B 파라미터`;
    }
  };
  
  /**
   * 기본 모델 크기 추정값들 (파라미터 수)
   */
  export const DEFAULT_MODEL_SIZES = {
    'small': 1000000,      // 1M parameters
    'medium': 10000000,    // 10M parameters  
    'large': 100000000,    // 100M parameters
    'xlarge': 1000000000   // 1B parameters
  };
  
  /**
   * 모델 정의가 분석되지 않았을 때 사용할 기본값 선택기
   */
  export const getDefaultModelSize = (modelType?: string): number => {
    const type = modelType?.toLowerCase();
    
    if (type?.includes('bert') || type?.includes('transformer')) {
      return DEFAULT_MODEL_SIZES.medium;
    } else if (type?.includes('resnet') || type?.includes('cnn')) {
      return DEFAULT_MODEL_SIZES.small;
    } else if (type?.includes('gpt') || type?.includes('llm')) {
      return DEFAULT_MODEL_SIZES.large;
    }
    
    return DEFAULT_MODEL_SIZES.small; // 기본값
  };