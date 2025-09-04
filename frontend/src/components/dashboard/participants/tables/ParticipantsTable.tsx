import { Edit, Trash2, CheckCircle, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Participant } from "@/types/participant";
import { getStatusBadge } from "../utils";

interface ParticipantsTableProps {
  participants: Participant[];
  selectedParticipant: Participant | null;
  onSelectParticipant: (participant: Participant) => void;
  onEditParticipant: (participant: Participant) => void;
  onViewVMs: (participant: Participant) => void;
  onHealthCheck: (participant: Participant) => void;
  onDeleteParticipant: (id: string) => void;
}

export function ParticipantsTable({
  participants,
  selectedParticipant,
  onSelectParticipant,
  onEditParticipant,
  onViewVMs,
  onHealthCheck,
  onDeleteParticipant,
}: ParticipantsTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>이름</TableHead>
          <TableHead>리전</TableHead>
          <TableHead>상태</TableHead>
          <TableHead>생성일</TableHead>
          <TableHead>액션</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {participants.map((participant) => (
          <TableRow
            key={participant.id}
            className={`cursor-pointer hover:bg-muted/50 ${
              selectedParticipant?.id === participant.id ? "bg-muted" : ""
            }`}
            onClick={() => onSelectParticipant(participant)}
          >
            <TableCell className="font-medium">{participant.name}</TableCell>
            <TableCell>
              {participant.region ? (
                <span className="text-sm bg-muted px-2 py-1 rounded">
                  {participant.region}
                </span>
              ) : (
                <span className="text-muted-foreground text-sm">-</span>
              )}
            </TableCell>
            <TableCell>{getStatusBadge(participant.status)}</TableCell>
            <TableCell>
              {new Date(participant.created_at).toLocaleDateString()}
            </TableCell>
            <TableCell>
              <div
                className="flex space-x-1"
                onClick={(e) => e.stopPropagation()}
              >
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onEditParticipant(participant)}
                  title="편집"
                >
                  <Edit className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onViewVMs(participant)}
                  title="VM 목록 보기"
                >
                  <Server className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onHealthCheck(participant)}
                  title="헬스체크"
                >
                  <CheckCircle className="h-4 w-4" />
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant="outline" size="sm" title="삭제">
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>클러스터 삭제</AlertDialogTitle>
                      <AlertDialogDescription>
                        이 클러스터를 삭제하시겠습니까? 이 작업은 되돌릴 수
                        없습니다.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>취소</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => onDeleteParticipant(participant.id)}
                      >
                        삭제
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </TableCell>
          </TableRow>
        ))}
        {participants.length === 0 && (
          <TableRow>
            <TableCell colSpan={5} className="text-center py-8">
              클러스터가 없습니다. 새 클러스터를 추가해보세요.
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
}
