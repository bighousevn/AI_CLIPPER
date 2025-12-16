import { useQuery } from "@tanstack/react-query";
import { getClips } from "~/services/clipService";

export function useClips() {
    return useQuery({ queryKey: ["clips"], queryFn: getClips });
}
