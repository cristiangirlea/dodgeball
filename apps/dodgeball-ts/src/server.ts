import express, { Request, Response } from "express";
import multer, { Multer } from "multer";
import { runSimulationTS } from "./client";
import { parseInput } from "./parser";

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

            const parsed = parseInput(text);

            const result = await runSimulationTS(parsed);

            // Output format
            const out = `${result.throws} ${result.lastPlayer}`;

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
