// components/dashboard/federated-learning/components/StatusBadge.tsx
import { Badge } from "@/components/ui/badge";

interface StatusBadgeProps {
  status: string;
}

export const StatusBadge = ({ status }: StatusBadgeProps) => {
  switch (status) {
    case "완료":
      return <Badge className="bg-green-500">완료</Badge>;
    case "진행중":
      return <Badge className="bg-blue-500">진행중</Badge>;
    case "대기중":
      return <Badge className="bg-yellow-500">대기중</Badge>;
    default:
      return <Badge>{status}</Badge>;
  }
};