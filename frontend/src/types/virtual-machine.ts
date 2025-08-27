// 가상머신 관련 타입들

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

// OpenStack VM 인스턴스 인터페이스
export interface OpenStackVMInstance {
  id: string;
  instance_id: string;
  name: string;
  status: string;
  flavor: {
    id: string;
    name: string;
    vcpus: number;
    ram: number; // MB 단위
    disk: number; // GB 단위
  };
  addresses: {
    [networkName: string]: Array<{
      addr: string;
      type: string;
    }>;
  };
  "OS-EXT-STS:power_state": number;
}

export interface SelectedVM {
  id: string;
  participant_id: string;
  name: string;
  instance_id: string;
  status: string;
  flavor_id: string;
  flavor_name: string;
  vcpus: number;
  ram: number;
  disk: number;
  ip_addresses: string;
  created_at: string;
  updated_at: string;
}

export interface OptimalVM {
  data: {
    selected_vm: SelectedVM;
    selection_reason: string;
    candidate_count: number;
  };
  message: string;
  success: boolean;
}

// 가상머신 정보 인터페이스 (DB에 저장된 정보)
export interface VirtualMachine {
  id: string;
  participant_id: string;
  name: string;
  instance_id: string;
  status: string;
  ip_address?: string;
  private_ip?: string;
  flavor_id?: string;
  image_id?: string;
  vcpus: number;
  ram: number;
  disk: number;
  last_health_check?: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_in_bytes: number;
  network_out_bytes: number;
  current_task_id?: string;
  task_assigned_at?: string;
  last_task_completed_at?: string;
  total_tasks_completed: number;
  metadata?: string;
  created_at: string;
  updated_at: string;
}
