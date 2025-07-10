// 참여자 관련 타입들

// 연합학습 참여자(OpenStack 클라우드) 인터페이스
export interface Participant {
	id: string;
	user_id: number;
	name: string;
	status: "active" | "inactive" | "busy" | "error" | "pending";
	metadata?: string;

	// OpenStack 클라우드 관련 필드
	openstack_endpoint?: string;
	openstack_region?: string;
	openstack_app_credential_id?: string;
	openstack_app_credential_secret?: string;

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
	openstack_region?: string;
	openstack_app_credential_id?: string;
	openstack_app_credential_secret?: string;
}
