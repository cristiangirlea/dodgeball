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
        return new Promise((resolve, reject) => {
            this.client[method](data, (err: grpc.ServiceError | null, response: any) => {
                if (err) return reject(err);

                // ts-proto expects Uint8Array
                resolve(new Uint8Array(response));
            });
        });
    }
}
