// components/dashboard/federated-learning/hooks/useParticipants.ts
import { useState, useCallback } from "react";
import { toast } from "sonner";
import { getAvailableParticipants } from "@/api/participants";
import { Participant } from "@/types/participant";

export const useParticipants = () => {
  const [participants, setParticipants] = useState<Participant[]>([]);

  const fetchParticipants = useCallback(async () => {
    try {
      const participantList = await getAvailableParticipants();
      setParticipants(participantList);
    } catch (error) {
      toast.error("참여자 목록 조회에 실패했습니다: " + error);
      setParticipants([]);
    }
  }, []);

  return {
    participants,
    fetchParticipants,
  };
};