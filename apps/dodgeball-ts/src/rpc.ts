import * as grpc from "@grpc/grpc-js";

// A minimal Rpc transport that sends/receives raw protobuf bytes over gRPC.
// It uses grpc-js low-level makeUnaryRequest with identity (Buffer) serializers,
// matching what ts-proto expects (bytes in, bytes out).
export class GrpcTransport {
    private client: grpc.Client;
    private service: string;

    constructor(address: string, _protoPath: string, fullServiceName: string) {
        // We keep the constructor signature for compatibility, but we don't
        // need the .proto at runtime because we operate on raw bytes.
        this.client = new grpc.Client(address, grpc.credentials.createInsecure());
        this.service = fullServiceName; // e.g. "dodgeball.DodgeballService"
    }

    request(service: string, method: string, data: Uint8Array): Promise<Uint8Array> {
        const isDev = process.env.NODE_ENV === "development";
        if (isDev) {
            try {
                const max = 64;
                const len = data?.length ?? 0;
                const slice = data ? data.slice(0, Math.min(max, len)) : new Uint8Array();
                const hex = Array.from(slice).map((b) => b.toString(16).padStart(2, "0")).join(" ");
                console.debug(`[gRPC] → ${service}.${method} payload: ${len} bytes${len > max ? ` (showing first ${max})` : ""}`);
                console.debug(`[gRPC] → ${service}.${method} hex: ${hex}${len > max ? " …" : ""}`);
            } catch {
                // best-effort logging only
            }
        }

        // Fully qualified gRPC method path: /package.Service/Method
        const path = `/${service}/${method}`;

        return new Promise((resolve, reject) => {
            this.client.makeUnaryRequest(
                path,
                // serialize: Uint8Array -> Buffer (identity at message level)
                (arg: Uint8Array) => Buffer.from(arg),
                // deserialize: Buffer -> Uint8Array (identity back)
                (data: Buffer) => new Uint8Array(data),
                data,
                (err: grpc.ServiceError | null, response?: Uint8Array) => {
                    if (err) {
                        if (isDev) {
                            console.debug(`[gRPC] ← ${service}.${method} error:`, err);
                        }
                        return reject(err);
                    }

                    const out = response ?? new Uint8Array();

                    if (isDev) {
                        try {
                            const max = 64;
                            const len = out.length;
                            const slice = out.slice(0, Math.min(max, len));
                            const hex = Array.from(slice).map((b) => b.toString(16).padStart(2, "0")).join(" ");
                            console.debug(`[gRPC] ← ${service}.${method} response: ${len} bytes${len > max ? ` (showing first ${max})` : ""}`);
                            console.debug(`[gRPC] ← ${service}.${method} hex: ${hex}${len > max ? " …" : ""}`);
                        } catch {
                            // ignore logging errors
                        }
                    }

                    resolve(out);
                }
            );
        });
    }
}
