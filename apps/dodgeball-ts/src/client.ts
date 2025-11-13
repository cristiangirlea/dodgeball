import { GrpcTransport } from "./rpc";
import {
    DodgeballServiceClientImpl,
    SimulationInput,
    SimulationResult
} from "../proto_gen/dodgeball";

const PROTO_PATH = "../proto/dodgeball.proto";

export async function runSimulationTS(
    input: SimulationInput,
    address = "127.0.0.1:50051"
): Promise<SimulationResult> {

    const transport = new GrpcTransport(
        address,
        PROTO_PATH,
        "dodgeball.DodgeballService"
    );

    const client = new DodgeballServiceClientImpl(transport);

    if (process.env.NODE_ENV === "development") {
        try {
            console.debug(`[gRPC] → dodgeball.DodgeballService.RunSimulation @ ${address}`);
            console.debug(`[gRPC] → request JSON:`, input);
        } catch {
            // best-effort logging only
        }
    }

    return client.RunSimulation(input);
}
