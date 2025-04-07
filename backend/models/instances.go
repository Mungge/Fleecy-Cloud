package models

type Instance struct {
	Name 	   string  `json:"name"` // name of instance
	VCPU	  int     `json:"vcpu"` // number of virtual CPUs
	RAMMB	int     `json:"ram_mb"` // amount of RAM in MB
	Family	string  `json:"family"` // ex: AWS t3, GCP e2
}

func RecommendInstance(est ResourceEstimate) []Instance{
	candidates := []Instance{
		{Name: "t3.medium", VCPU: 2, RAMMB: 4096, Family: "AWS t3"},
		{Name: "t3.large", VCPU: 2, RAMMB: 8192, Family: "AWS t3"},
		{Name: "m5.large", VCPU: 2, RAMMB: 8192, Family: "AWS m5"},
		{Name: "m5.xlarge", VCPU: 4, RAMMB: 16384, Family: "AWS m5"},
		{Name: "c5.large", VCPU: 2, RAMMB: 4096, Family: "AWS c5"},
		{Name: "e2-medium", VCPU: 2, RAMMB: 4096, Family: "GCP e2"},
		{Name: "e2-standard-2", VCPU: 2, RAMMB: 8192, Family: "GCP e2"},
	}

	var result []Instance
	for _, inst := range candidates{
		// 조건: RAM >= 예측치 AND vCPU tn x 50% >= CPU 사용량 예측
		if inst.RAMMB >= est.RAMMB && (inst.VCPU*50 >= est.CPUPercent) {
			result = append(result, inst)
		}
	}
	if len(result) == 0 {
		// 추천 가능한 인스턴스가 없을 경우 fallback
		return []Instance{
			{Name: "custom-needed", VCPU: 8, RAMMB: 32768, Family: "Custom"},
		}
	}
	return result
}