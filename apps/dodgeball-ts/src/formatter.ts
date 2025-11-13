import { SimulationResult } from "../proto_gen/dodgeball";

export function formatOutput(result: SimulationResult): string {
    return `${result.throws} ${result.lastPlayer}`;
}
