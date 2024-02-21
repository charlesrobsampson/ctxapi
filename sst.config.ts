import { SSTConfig } from "sst";
import { Api, Table } from "sst/constructs";

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
      const api = new Api(stack, "api", {
        defaults: {
          function: {
            bind: [mainTable]
          }
        },
        routes: {
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
        },
      });
      stack.addOutputs({
        ApiEndpoint: api.url,
      });
    });
  },
} satisfies SSTConfig;
