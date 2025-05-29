"use client";

import { useEffect, useState } from "react";
import { useToast } from "./use-toast";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

export function Toaster() {
	const { toasts } = useToast();

	return (
		<div className="fixed top-0 z-[100] flex flex-col items-end justify-start gap-2 w-full max-w-md p-4 pointer-events-none">
			{toasts.map((toast, i) => (
				<Toast key={i} {...toast} />
			))}
		</div>
	);
}

function Toast({
	title,
	description,
	variant = "default",
}: {
	title?: string;
	description?: string;
	variant?: "default" | "destructive" | "success";
}) {
	const [isVisible, setIsVisible] = useState(false);

	useEffect(() => {
		setIsVisible(true);

		const timer = setTimeout(() => {
			setIsVisible(false);
		}, 2500);

		return () => clearTimeout(timer);
	}, []);

	return (
		<div
			className={cn(
				"bg-white rounded-lg shadow-lg border p-4 pointer-events-auto w-full max-w-sm transform transition-all duration-300 ease-out",
				isVisible ? "translate-y-0 opacity-100" : "translate-y-2 opacity-0",
				variant === "destructive" && "border-red-500",
				variant === "success" && "border-green-500"
			)}
			role="alert"
		>
			<div className="flex items-start">
				<div className="flex-1">
					{title && <h5 className="font-medium text-gray-900">{title}</h5>}
					{description && (
						<p className="text-sm text-gray-600 mt-1">{description}</p>
					)}
				</div>
				<button
					onClick={() => setIsVisible(false)}
					className="inline-flex text-gray-500 hover:text-gray-700"
				>
					<X className="h-4 w-4" />
				</button>
			</div>
		</div>
	);
}
