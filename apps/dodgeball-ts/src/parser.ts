import { SimulationInput } from "../proto_gen/dodgeball";

function dirStrToCode(s: string): number {
    switch (s.trim().toUpperCase()) {
        case "N": return 0;
        case "NE": return 1;
        case "E": return 2;
        case "SE": return 3;
        case "S": return 4;
        case "SW": return 5;
        case "W": return 6;
        case "NW": return 7;
        default:
            throw new Error(`Unknown direction: ${s}`);
    }
}

function parseSingleCaseFromTokens(tokens: string[], start: number): { input: SimulationInput; next: number } {
    let i = start;
    const n = parseInt(tokens[i++], 10);
    if (!Number.isFinite(n) || n < 0) {
        throw new Error(`Invalid number of players at token ${start}: ${tokens[start]}`);
    }

    const players: SimulationInput["players"] = [];
    for (let j = 0; j < n; j++) {
        const xTok = tokens[i++];
        const yTok = tokens[i++];
        if (xTok === undefined || yTok === undefined) {
            throw new Error(`Unexpected EOF while reading player ${j + 1}/${n} at token index ${i - 1}`);
        }
        const x = Number(xTok);
        const y = Number(yTok);
        if (!Number.isFinite(x) || !Number.isFinite(y)) {
            throw new Error(`Invalid coordinate at player ${j + 1}: (${xTok}, ${yTok})`);
        }
        players.push({ x, y, alive: true });
    }

    const directionStr = tokens[i++];
    if (directionStr === undefined) {
        throw new Error(`Unexpected EOF while reading direction after ${n} players`);
    }
    const startIndex1BasedTok = tokens[i++];
    if (startIndex1BasedTok === undefined) {
        throw new Error(`Unexpected EOF while reading starting index`);
    }
    const startIndex1Based = parseInt(startIndex1BasedTok, 10);
    if (!Number.isFinite(startIndex1Based)) {
        throw new Error(`Invalid starting index: ${startIndex1BasedTok}`);
    }

    return {
        input: {
            players,
            startDirection: dirStrToCode(directionStr),
            startIndex: startIndex1Based - 1,
        },
        next: i,
    };
}

export function parseInputsFromText(text: string, opts?: { firstOnly?: boolean }): SimulationInput[] {
    const tokens = text.trim().split(/\s+/);
    const firstOnly = !!opts?.firstOnly;

    // Strategy A: assume a leading T (number of test cases). Fallback to single-case.
    const tryWithLeadingT = (): SimulationInput[] => {
        let i = 0;
        const T = parseInt(tokens[i++], 10);
        if (!Number.isFinite(T) || T <= 0) {
            throw new Error(`Invalid test case count T at token 0: ${tokens[0]}`);
        }
        const inputs: SimulationInput[] = [];
        for (let t = 0; t < T; t++) {
            const { input, next } = parseSingleCaseFromTokens(tokens, i);
            inputs.push(input);
            i = next;
            if (firstOnly) break;
        }
        return inputs;
    };

    // Try parsing as multi-case (T present). If it fails, parse as single-case from start.
    try {
        const inputs = tryWithLeadingT();
        if (!firstOnly || inputs.length > 0) {
            return firstOnly ? inputs.slice(0, 1) : inputs;
        }
    } catch {
        // ignore, will try single-case
    }

    // Single case from the beginning (no leading T)
    const { input } = parseSingleCaseFromTokens(tokens, 0);
    return [input];
}

export function parseInputsFromJson(text: string): SimulationInput[] {
    const raw = JSON.parse(text);
    const toInput = (obj: any): SimulationInput => {
        if (!obj || !Array.isArray(obj.players)) {
            throw new Error("JSON: missing 'players' array");
        }
        const players: SimulationInput["players"] = obj.players.map((p: any, idx: number) => {
            if (!Array.isArray(p) || p.length !== 2) {
                throw new Error(`JSON: player at index ${idx} must be a [x,y] array`);
            }
            const x = Number(p[0]);
            const y = Number(p[1]);
            if (!Number.isFinite(x) || !Number.isFinite(y)) {
                throw new Error(`JSON: invalid coordinate at player ${idx + 1}: (${p[0]}, ${p[1]})`);
            }
            return { x, y, alive: true };
        });
        const dir = obj.startingDirection ?? obj.direction ?? obj.dir ?? obj.startDirection;
        if (typeof dir !== "string") {
            throw new Error("JSON: missing 'startingDirection' (string)");
        }
        const s1 = obj.startingPlayer ?? obj.start ?? obj.s ?? obj.startIndex;
        if (!Number.isFinite(Number(s1))) {
            throw new Error("JSON: missing 'startingPlayer' (1-based integer)");
        }
        return {
            players,
            startDirection: dirStrToCode(String(dir)),
            startIndex: Number(s1) - 1,
        };
    };

    if (Array.isArray(raw)) {
        return raw.map(toInput);
    }
    return [toInput(raw)];
}

export function autoParseInputs(text: string): SimulationInput[] {
    const trimmed = text.trim();
    if (trimmed.startsWith("[") || trimmed.startsWith("{")) {
        return parseInputsFromJson(text);
    }
    return parseInputsFromText(text, { firstOnly: false });
}

// Backward compatibility: parse only the first case from plain text
export function parseInput(text: string): SimulationInput {
    return parseInputsFromText(text, { firstOnly: true })[0];
}
