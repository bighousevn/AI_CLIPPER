import { z } from "zod";

export const ClipConfigSchema = z.object({
    prompt: z
        .string()
        .min(5, "Prompt must be at least 5 characters")
        .max(300, "Prompt too long"),

    numberOfClips: z
        .number()
        .min(1, "At least 1 clip")
        .max(10, "Max 10 clips"),

    duration: z
        .number()
        .min(3, "Min duration 3s")
        .max(60, "Max duration 60s"),

    aspectRatio: z.enum(["9:16", "16:9", "1:1"]),
});
