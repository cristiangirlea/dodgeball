import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";

export class GrpcTransport {
    private client: any;

    constructor(address: string, protoPath: string, fullServiceName: string) {
        const [packageName, serviceName] = fullServiceName.split(".");

        const pkgDef = protoLoader.loadSync(protoPath, {
            longs: String,
            enums: String,
            defaults: false,
        });

        const grpcObj = grpc.loadPackageDefinition(pkgDef) as any;

        if (!grpcObj[packageName] || !grpcObj[packageName][serviceName]) {
            throw new Error(
                `Service ${packageName}.${serviceName} not found in proto`
            );
        }

        this.client = new grpcObj[packageName][serviceName](
            address,
            grpc.credentials.createInsecure()
        );
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
        return new Promise((resolve, reject) => {
            this.client[method](data, (err: grpc.ServiceError | null, response: any) => {
                if (err) {
                    if (isDev) {
                        console.debug(`[gRPC] ← ${service}.${method} error:`, err);
                    }
                    return reject(err);
                }

                // ts-proto expects Uint8Array
                const out = new Uint8Array(response);

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
            });
        });
    }
}
