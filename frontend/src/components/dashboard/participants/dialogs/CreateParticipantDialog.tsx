import { Plus } from "lucide-react";
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
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ParticipantFormData } from "@/hooks/participants/useParticipantForm";
import { UseFormReturn } from "react-hook-form";

interface CreateParticipantDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	form: UseFormReturn<ParticipantFormData>;
	configFile: File | null;
	onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
	onSubmit: (data: ParticipantFormData) => void;
	onClose: () => void;
}

export function CreateParticipantDialog({
	open,
	onOpenChange,
	form,
	configFile,
	onFileChange,
	onSubmit,
	onClose,
}: CreateParticipantDialogProps) {
	return (
		<Dialog
			open={open}
			onOpenChange={(isOpen) => {
				if (!isOpen) {
					onClose();
				}
				onOpenChange(isOpen);
			}}
		>
			<DialogTrigger asChild>
				<Button>
					<Plus className="mr-2 h-4 w-4" />
					클러스터 추가
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-2xl">
				<DialogHeader>
					<DialogTitle>클러스터 추가</DialogTitle>
					<DialogDescription>
						새로운 클러스터 정보를 입력하세요.
					</DialogDescription>
				</DialogHeader>

				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>이름</FormLabel>
									<FormControl>
										<Input placeholder="참여자 이름" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="metadata"
							render={({ field }) => (
								<FormItem>
									<FormLabel>메타데이터</FormLabel>
									<FormControl>
										<Input placeholder="추가 정보" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						{/* OpenStack 설정 YAML 파일 업로드 */}
						<div className="space-y-4 border-t pt-4">
							<div>
								<Label className="text-base font-semibold">
									OpenStack 설정
								</Label>
								<p className="text-sm text-muted-foreground mt-1">
									OpenStack 클러스터 설정이 포함된 YAML 파일을 업로드하세요.
								</p>
							</div>

							<div className="space-y-2">
								<Label htmlFor="config-file">설정 파일 (*.yaml, *.yml)</Label>
								<Input
									id="config-file"
									type="file"
									accept=".yaml,.yml"
									onChange={onFileChange}
									className="cursor-pointer"
								/>
								{configFile && (
									<div className="text-sm text-green-600">
										선택된 파일: {configFile.name} (
										{Math.round(configFile.size / 1024)} KB)
									</div>
								)}
							</div>
						</div>

						<DialogFooter>
							<Button type="submit">클러스터 추가</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
