import express, { Request, Response } from "express";
import multer, { Multer } from "multer";
import { runSimulationTS } from "./client";
import { autoParseInputs } from "./parser";

const app = express();
const upload: Multer = multer({ storage: multer.memoryStorage() });

const PORT = 3000;

app.post(
    "/simulate",
    upload.single("input"),
    async (req: Request, res: Response) => {
        try {
            if (!req.file) {
                return res.status(400).send("No input file uploaded");
            }

            const text = req.file.buffer.toString("utf8");

            // Supports both text (with optional leading T and multiple cases)
            // and JSON (single object or array) inputs.
            const inputs = autoParseInputs(text);

            const lines: string[] = [];
            for (const input of inputs) {
                const result = await runSimulationTS(input);
                // Output format (match sample .out): throws and 1-based last player index
                lines.push(`${result.throws} ${result.lastPlayer + 1}`);
            }

            const out = lines.join("\n");
            return res.type("text/plain").send(out);

        } catch (err: any) {
            console.error(err);
            return res.status(500).send(err.message ?? "Internal error");
        }
    }
);

app.listen(PORT, () => {
    console.log(`Dodgeball TS server running on http://localhost:${PORT}`);
});
