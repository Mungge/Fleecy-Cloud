import { useState } from "react";

interface ToastProps {
	title?: string;
	description?: string;
	variant?: "default" | "destructive" | "success";
}

export function useToast() {
	const [toasts, setToasts] = useState<ToastProps[]>([]);

	const toast = (props: ToastProps) => {
		setToasts((prev) => [...prev, props]);

		// 3초 후에 토스트 제거
		setTimeout(() => {
			setToasts((prev) => prev.filter((_, i) => i !== 0));
		}, 3000);
	};

	return { toast, toasts };
}
