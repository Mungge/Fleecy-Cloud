// components/dashboard/federated-learning/components/CreateJobDialog.tsx
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Slider } from "@/components/ui/slider";
import { Participant } from "@/types/participant";
import { AGGREGATION_ALGORITHMS, MODEL_TYPES, SUPPORTED_FILE_FORMATS } from "../constants";
import { CreateJobFormHookReturn } from "../types";

interface CreateJobDialogProps {
  participants: Participant[];
  formHook: CreateJobFormHookReturn;
}

export const CreateJobDialog = ({ participants, formHook }: CreateJobDialogProps) => {
  const { form, modelFile, setModelFile, isDialogOpen, openDialog, closeDialog, handleSubmit } = formHook;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setModelFile(e.target.files[0]);
    }
  };

  return (
    <Dialog
      open={isDialogOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeDialog();
        }
      }}
    >
      <DialogTrigger asChild>
        <Button className="ml-auto" onClick={openDialog}>
          <Plus className="mr-2 h-4 w-4" />
          연합학습 생성
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>연합학습 생성</DialogTitle>
          <DialogDescription>
            새로운 연합학습 작업에 필요한 정보를 입력하세요.
          </DialogDescription>
        </DialogHeader>

        {/* Progress Steps */}
        <div className="w-full py-4">
          <div className="flex items-center justify-between">
            {/* Step 1: 정보 입력 */}
            <div className="flex flex-col items-center">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-blue-500 text-white text-sm font-medium">
                1
              </div>
              <span className="mt-2 text-sm font-medium text-blue-600">
                정보 입력
              </span>
            </div>

            {/* Connector Line */}
            <div className="flex-1 h-0.5 bg-gray-200 mx-4"></div>

            {/* Step 2: 집계자 생성 */}
            <div className="flex flex-col items-center">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-200 text-gray-400 text-sm font-medium">
                2
              </div>
              <span className="mt-2 text-sm text-gray-400">
                집계자 생성
              </span>
            </div>

            {/* Connector Line */}
            <div className="flex-1 h-0.5 bg-gray-200 mx-4"></div>

            {/* Step 3: 연합학습 생성 */}
            <div className="flex flex-col items-center">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-200 text-gray-400 text-sm font-medium">
                3
              </div>
              <span className="mt-2 text-sm text-gray-400">
                연합학습 생성
              </span>
            </div>
          </div>
        </div>

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-6"
          >
            <Tabs defaultValue="basic" className="w-full">
              <TabsList className="grid grid-cols-3 mb-4">
                <TabsTrigger value="basic">기본 정보</TabsTrigger>
                <TabsTrigger value="model">모델 설정</TabsTrigger>
                <TabsTrigger value="participants">참여자 설정</TabsTrigger>
              </TabsList>

              <TabsContent value="basic" className="space-y-4">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>이름</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="연합학습 작업 이름"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>설명</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="작업에 대한 간략한 설명"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="algorithm"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>집계 알고리즘</FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        defaultValue={field.value}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="집계 알고리즘 선택" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {AGGREGATION_ALGORITHMS.map((algo) => (
                            <SelectItem key={algo.id} value={algo.id}>
                              {algo.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormDescription>
                        클라이언트 모델을 집계하는 알고리즘입니다.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="rounds"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>라운드 수: {field.value}</FormLabel>
                      <FormControl>
                        <Slider
                          defaultValue={[field.value]}
                          min={1}
                          max={100}
                          step={1}
                          onValueChange={(vals) => field.onChange(vals[0])}
                        />
                      </FormControl>
                      <FormDescription>
                        연합학습이 수행될 라운드 수를 설정하세요.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </TabsContent>

              <TabsContent value="model" className="space-y-4">
                <FormField
                  control={form.control}
                  name="modelType"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>모델 유형</FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        defaultValue={field.value}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="모델 유형 선택" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {MODEL_TYPES.map((type) => (
                            <SelectItem key={type.id} value={type.id}>
                              {type.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormDescription>
                        연합학습에 사용될 모델의 유형을 선택하세요.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="space-y-2">
                  <Label>모델 파일 업로드</Label>
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Input
                      type="file"
                      accept={SUPPORTED_FILE_FORMATS}
                      onChange={handleFileChange}
                    />
                    <p className="text-sm text-muted-foreground">
                      지원 형식: .h5, .pb, .pt, .pth, .onnx, .pkl
                    </p>
                  </div>
                  {modelFile && (
                    <div className="text-sm">
                      선택된 파일: {modelFile.name} (
                      {Math.round(modelFile.size / 1024)} KB)
                    </div>
                  )}
                </div>
              </TabsContent>

              <TabsContent value="participants" className="space-y-4">
                <FormField
                  control={form.control}
                  name="participants"
                  render={() => (
                    <FormItem>
                      <div className="mb-4">
                        <FormLabel className="text-base">
                          참여자 선택
                        </FormLabel>
                        <FormDescription>
                          연합학습에 참여할 참여자를 선택하세요. 최소 1개
                          이상의 참여자가 필요합니다.
                        </FormDescription>
                        <FormMessage />
                      </div>
                      <div className="space-y-4">
                        {participants.map((participant) => (
                          <FormField
                            key={participant.id}
                            control={form.control}
                            name="participants"
                            render={({ field }) => {
                              return (
                                <FormItem
                                  key={participant.id}
                                  className="flex flex-row items-start space-x-3 space-y-0"
                                >
                                  <FormControl>
                                    <Checkbox
                                      disabled={
                                        participant.status === "inactive"
                                      }
                                      checked={field.value?.includes(
                                        participant.id
                                      )}
                                      onCheckedChange={(checked) => {
                                        return checked
                                          ? field.onChange([
                                              ...field.value,
                                              participant.id,
                                            ])
                                          : field.onChange(
                                              field.value?.filter(
                                                (value: string) =>
                                                  value !== participant.id
                                              )
                                            );
                                      }}
                                    />
                                  </FormControl>
                                  <div className="space-y-1 leading-none">
                                    <FormLabel
                                      className={
                                        participant.status === "inactive"
                                          ? "text-muted-foreground"
                                          : ""
                                      }
                                    >
                                      {participant.name}
                                    </FormLabel>
                                    <div className="text-xs text-muted-foreground">
                                      {participant.openstack_endpoint ||
                                        "OpenStack 엔드포인트 정보 없음"}
                                      <span
                                        className={
                                          participant.status === "active"
                                            ? "text-green-500 ml-1"
                                            : "text-red-500 ml-1"
                                        }
                                      >
                                        {participant.status === "active"
                                          ? "활성"
                                          : "비활성"}
                                      </span>
                                    </div>
                                  </div>
                                </FormItem>
                              );
                            }}
                          />
                        ))}
                        {participants.length === 0 && (
                          <div className="text-center py-4 text-muted-foreground">
                            사용 가능한 참여자가 없습니다.
                            <br />
                            참여자 정보를 먼저 설정해주세요.
                          </div>
                        )}
                      </div>
                    </FormItem>
                  )}
                />
              </TabsContent>
            </Tabs>

            <DialogFooter>
              <Button type="submit">다음: 집계자 생성</Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};