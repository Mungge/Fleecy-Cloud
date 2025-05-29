export interface FederatedLearningJob {
	id: string;
	user_id: number;
	name: string;
	description?: string;
	status: string;
	participants: number;
	completed_at: string | null;
	accuracy: number;
	recall: number;
	precision: number;
	f1score: number;
	rounds: number;
	algorithm: string;
	model_type: string;
	created_at: string;
	updated_at: string;
}

export interface CloudParticipant {
	id: string;
	name: string;
	region: string;
	status: "active" | "inactive";
}

// 연합학습 참여자(OpenStack 클라우드) 인터페이스
export interface Participant {
	id: string;
	user_id: number;
	name: string;
	status: "active" | "inactive" | "busy" | "error" | "pending";
	metadata?: string;

	// OpenStack 클라우드 관련 필드
	openstack_endpoint?: string;
	openstack_username?: string;
	openstack_password?: string;
	openstack_project_name?: string;
	openstack_domain_name?: string;
	openstack_region?: string;

	// VM 모니터링 관련 필드
	vm_status?: string;
	last_health_check?: string;
	cpu_usage?: number;
	memory_usage?: number;
	disk_usage?: number;
	network_in_bytes?: number;
	network_out_bytes?: number;

	// 연합학습 관련 필드
	current_task_id?: string;
	task_assigned_at?: string;
	last_task_completed_at?: string;
	total_tasks_completed?: number;

	created_at: string;
	updated_at: string;
}

// 참여자 생성 요청 인터페이스
export interface CreateParticipantRequest {
	name: string;
	metadata?: string;

	// OpenStack 클라우드 관련 필드
	openstack_endpoint?: string;
	openstack_username?: string;
	openstack_password?: string;
	openstack_project_name?: string;
	openstack_domain_name?: string;
	openstack_region?: string;
}

// VM 모니터링 정보 인터페이스
export interface VMMonitoringInfo {
	instance_id: string;
	status: string;
	availability_zone: string;
	host: string;
	created_at: string;
	updated_at: string;
	cpu_usage: number;
	memory_usage: number;
	disk_usage: number;
	network_in: number;
	network_out: number;

	// 연합 학습 관련 필드
	federated_learning_status?: string;
	current_task_id?: string;
	task_progress?: number;
	last_training_time?: string;
}

// VM 헬스체크 결과 인터페이스
export interface VMHealthCheckResult {
	healthy: boolean;
	status: string;
	message: string;
	checked_at: string;
	response_time_ms: number;
}
