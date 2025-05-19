package models

type UserInput struct {
	MaxClients 			int `json:"max_clients"`
	ModelSizeMB 		float64 `json:"avg_model_size_mb"`
	FLOPs 				float64 `json:"flops"`
	UploadFreqMin 		int `json:"upload_freq_min"`
	AggregationType 	string `json:"aggregation_type"`
}

type ResourceEstimate struct{
	RAMGB		int `json:"ram_gb"` // 예측된 RAM 사용량
	CPUPercent 	int `json:"cpu_percent"` // 예측된 CPU 사용량
	NetMBps 	int `json:"net_mb_per_second"`// 
}