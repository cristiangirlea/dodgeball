import fs from "fs";
import path from "path";
import { runSimulationTS } from "../src/client";

function parseSample(file) {
    const lines = fs.readFileSync(file, "utf8").trim().split(/\s+/);

    let idx = 0;

    const numPlayers = parseInt(lines[idx++], 10);
    const startDirection = parseInt(lines[idx++], 10);

    const players = [];
    for (let i = 0; i < numPlayers; i++) {
        const x = parseInt(lines[idx++], 10);
        const y = parseInt(lines[idx++], 10);
        players.push({ x, y, alive: true });
    }

    const directionStr = lines[idx++]; // "NW", "SE", etc.
    const startIndex = parseInt(lines[idx++], 10);

    return {
        players,
        startDirection,
        startIndex,
    };
}

async function testSample(sampleName) {
    const inputFile = path.join(__dirname, "samples", `${sampleName}.in`);
    const outputFile = path.join(__dirname, "samples", `${sampleName}.out`);

    const expectedLines = fs.readFileSync(outputFile, "utf8").trim().split(/\s+/);

    const input = parseSample(inputFile);
    const result = await runSimulationTS(input);

    console.log(`=== ${sampleName} ===`);
    console.log("Expected:", expectedLines.join(" "));
    console.log("Received:", result.throws, result.lastPlayer);
}

async function main() {
    await testSample("sample1");
    await testSample("sample2");
    await testSample("sample3");
}

main();
