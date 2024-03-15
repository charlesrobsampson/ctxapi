import { SSTConfig } from "sst";
import { Api, Table } from "sst/constructs";
const version = process.env.npm_package_version || '';

export default {
  config(_input) {
    return {
      name: "ctxapi",
      region: "us-west-1",
    };
  },
  stacks(app) {
    app.setDefaultFunctionProps({
      runtime: "go",
    });
    app.stack(function Stack({ stack }) {
      const mainTable = new Table(stack, "context", {
        fields: {
          userId: "string",
          contextId: "string",
        },
        primaryIndex: {
          partitionKey: "userId",
          sortKey: "contextId"
        },
      });
      const queueTable = new Table(stack, "queue", {
        fields: {
          userId: "string",
          id: "string",
        },
        primaryIndex: {
          partitionKey: "userId",
          sortKey: "id"
        },
      });
      const api = new Api(stack, "api", {
        defaults: {
          function: {
            bind: [
              mainTable,
              queueTable,
            ],
          }
        },
        routes: {
          "GET /version": {
            function: {
              timeout: '30 seconds',
              handler: "functions/version/main.go",
              permissions: [],
              environment: {
                version,
              }
            }
          },
          "GET /context/{userId}": {
            function: {
              timeout: '30 seconds',
              handler: "functions/context/get/main.go",
              permissions: [
                mainTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
              }
            }
          },
          "GET /context/{userId}/list": {
            function: {
              timeout: '30 seconds',
              handler: "functions/context/list/main.go",
              permissions: [
                mainTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
              }
            }
          },
          "POST /context/{userId}": {
            function: {
              timeout: '30 seconds',
              handler: "functions/context/update/main.go",
              permissions: [
                mainTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
              }
            }
          },
          "POST /context/{userId}/{contextId}/close": {
            function: {
              timeout: '30 seconds',
              handler: "functions/context/close/main.go",
              permissions: [
                mainTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
              }
            }
          },
          "GET /queue/{userId}/{queueId}": {
            function: {
              timeout: '30 seconds',
              handler: "functions/queue/get/main.go",
              permissions: [
                mainTable,
                queueTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
                queueTableName: queueTable.tableName,
              }
            }
          },
          "GET /queue/{userId}": {
            function: {
              timeout: '30 seconds',
              handler: "functions/queue/list/main.go",
              permissions: [
                mainTable,
                queueTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
                queueTableName: queueTable.tableName,
              }
            }
          },
          "POST /queue/{userId}": {
            function: {
              timeout: '30 seconds',
              handler: "functions/queue/update/main.go",
              permissions: [
                mainTable,
                queueTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
                queueTableName: queueTable.tableName,
              }
            }
          },
          "POST /queue/{userId}/{queueId}/start": {
            function: {
              timeout: '30 seconds',
              handler: "functions/queue/start/main.go",
              permissions: [
                mainTable,
                queueTable,
              ],
              environment: {
                mainTableName: mainTable.tableName,
                queueTableName: queueTable.tableName,
              }
            }
          },
        },
      });
      stack.addOutputs({
        ApiEndpoint: api.url,
      });
    });
  },
} satisfies SSTConfig;
