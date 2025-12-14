import { z } from "zod";

export const ClipConfigSchema = z.object({
    prompt: z
        .string()
        .min(5, "Prompt must be at least 5 characters")
        .max(300, "Prompt too long"),

    clipCount: z
        .number()
        .min(1, "At least 1 clip")
        .max(10, "Max 10 clips"),
    aspectRatio: z.enum(["9:16", "16:9", "1:1", ""]),
    subtitle: z.boolean(),
});
