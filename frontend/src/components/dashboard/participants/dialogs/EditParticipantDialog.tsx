import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
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
import { Button } from "@/components/ui/button";
import { ParticipantFormData } from "@/hooks/participants/useParticipantForm";
import { UseFormReturn } from "react-hook-form";

interface EditParticipantDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	form: UseFormReturn<ParticipantFormData>;
	configFile: File | null;
	onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
	onSubmit: (data: ParticipantFormData) => void;
}

export function EditParticipantDialog({
	open,
	onOpenChange,
	form,
	configFile,
	onFileChange,
	onSubmit,
}: EditParticipantDialogProps) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-2xl">
				<DialogHeader>
					<DialogTitle>클러스터 수정</DialogTitle>
					<DialogDescription>클러스터 정보를 수정하세요.</DialogDescription>
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
										<Input placeholder="클러스터 이름" {...field} />
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

						{/* OpenStack 설정 업데이트 */}
						<div className="space-y-4 border-t pt-4">
							<h3 className="text-lg font-semibold">OpenStack 설정 업데이트</h3>
							<p className="text-sm text-muted-foreground">
								기존 설정을 유지하거나 새로운 YAML 파일로 업데이트할 수
								있습니다.
							</p>

							<div className="space-y-2">
								<Label htmlFor="edit-config-file">
									새 설정 파일 (선택사항)
								</Label>
								<Input
									id="edit-config-file"
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
								<p className="text-xs text-muted-foreground">
									파일을 선택하지 않으면 기존 설정이 유지됩니다.
								</p>
							</div>
						</div>

						<DialogFooter>
							<Button type="submit">클러스터 수정</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
