// components/dashboard/federated-learning/components/FederatedLearningList.tsx
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
  } from "@/components/ui/card";
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from "@/components/ui/table";
  import { ScrollArea } from "@/components/ui/scroll-area";
  import { StatusBadge } from "./StatusBadge";
  import { DeleteConfirmDialog } from "./DeleteConfirmDialog";
  import { FederatedLearningJob } from "@/types/federated-learning";
  
  interface FederatedLearningListProps {
    jobs: FederatedLearningJob[];
    isLoading: boolean;
    onJobSelect: (job: FederatedLearningJob) => void;
    onJobDelete: (id: string) => void;
  }
  
  export const FederatedLearningList = ({
    jobs,
    isLoading,
    onJobSelect,
    onJobDelete,
  }: FederatedLearningListProps) => {
    if (isLoading) {
      return (
        <Card className="md:col-span-2">
          <CardContent className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
          </CardContent>
        </Card>
      );
    }
  
    return (
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle>연합학습 목록</CardTitle>
          <CardDescription>
            연합학습 작업의 상태와 세부 정보를 확인하세요.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-[calc(100vh-320px)]">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>이름</TableHead>
                  <TableHead>상태</TableHead>
                  <TableHead>참여자</TableHead>
                  <TableHead>생성일</TableHead>
                  <TableHead>액션</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {jobs.map((job) => (
                  <TableRow
                    key={job.id}
                    className="cursor-pointer"
                    onClick={() => onJobSelect(job)}
                  >
                    <TableCell className="font-medium">{job.name}</TableCell>
                    <TableCell>
                      <StatusBadge status={job.status} />
                    </TableCell>
                    <TableCell>{job.participants}</TableCell>
                    <TableCell>{job.created_at}</TableCell>
                    <TableCell>
                      <DeleteConfirmDialog
                        onConfirm={() => onJobDelete(job.id)}
                        title="연합학습 삭제"
                        description="이 연합학습 작업을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다."
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>
    );
  };