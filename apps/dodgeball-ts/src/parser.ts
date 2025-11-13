import { SimulationInput } from "../proto_gen/dodgeball";

export function parseInput(text: string): SimulationInput {
    const tokens = text.trim().split(/\s+/);
    let i = 0;

    const numPlayers = parseInt(tokens[i++], 10);
    const startDirection = parseInt(tokens[i++], 10);

    const players = [];
    for (let j = 0; j < numPlayers; j++) {
        const x = parseInt(tokens[i++], 10);
        const y = parseInt(tokens[i++], 10);
        players.push({ x, y, alive: true });
    }

    const _directionStr = tokens[i++]; // we ignore this
    const startIndex = parseInt(tokens[i++], 10);

    return {
        players,
        startDirection,
        startIndex
    };
}
