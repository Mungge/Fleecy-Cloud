// components/dashboard/federated-learning/federated-learning-content.tsx
"use client";

import { useEffect, useRef } from "react";
import { useFederatedLearning } from "./hooks/useFederatedLearning";
import { useParticipants } from "./hooks/useParticipants";
import { useCreateJobForm } from "./hooks/useCreateJobForm";
import { FederatedLearningList } from "./components/FederatedLearningList";
import { FederatedLearningDetail } from "./components/FederatedLearningDetail";
import { CreateJobDialog } from "./components/CreateJobDialog";

const FederatedLearningContent = () => {
	const federatedLearning = useFederatedLearning();
	const participantsHook = useParticipants();
	const createFormHook = useCreateJobForm(participantsHook.participants);
	const initialized = useRef(false);

	useEffect(() => {
		if (initialized.current) return;

		let mounted = true;

		const fetchJobs = federatedLearning.fetchJobs;
		const fetchParticipants = participantsHook.fetchParticipants;

		const loadData = async () => {
			if (mounted) {
				await Promise.all([fetchJobs(), fetchParticipants()]);
				initialized.current = true;
			}
		};

		loadData();

		return () => {
			mounted = false;
		};
	}, [federatedLearning.fetchJobs, participantsHook.fetchParticipants]);

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex items-center justify-between">
				<div>
					<h2 className="text-3xl font-bold tracking-tight">연합학습</h2>
					<p className="text-muted-foreground">
						연합학습 작업을 생성하고 모니터링하세요.
					</p>
				</div>

				<CreateJobDialog
					participants={participantsHook.participants}
					formHook={createFormHook}
				/>
			</div>

			{/* Main Content Grid */}
			<div className="grid grid-cols-1 md:grid-cols-3 gap-6">
				<FederatedLearningList
					jobs={federatedLearning.jobs}
					isLoading={federatedLearning.isLoading}
					onJobSelect={federatedLearning.selectJob}
					onJobDelete={federatedLearning.deleteJob}
				/>

				<FederatedLearningDetail selectedJob={federatedLearning.selectedJob} />
			</div>
		</div>
	);
};

export default FederatedLearningContent;
