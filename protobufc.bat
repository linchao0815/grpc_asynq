set WORK_DIR=%cd%
set WORK_PROTO=%WORK_DIR%\api\v1
set INCLUDE_PROTO=%WORK_DIR%\api\gen
@rem del %WORK_DIR%\api\gen\v1 /S /Q
cd api/v1
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=%WORK_DIR%\api\gen\v1 --go_opt=paths=source_relative ./common.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=%WORK_DIR%\api\gen\v1 --go_opt=paths=source_relative --go-grpc_out=%WORK_DIR%\api\gen\v1 --go-grpc_opt=paths=source_relative --grpc-gateway_out=%WORK_DIR%\api\gen\v1 --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=%WORK_DIR%\api\gen\v1 --openapiv2_opt logtostderr=true ./serverA.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=paths=source_relative:%WORK_DIR%\api\gen\v1 --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./serverA_task.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=%WORK_DIR%\api\gen\v1 --go_opt=paths=source_relative --go-grpc_out=%WORK_DIR%\api\gen\v1 --go-grpc_opt=paths=source_relative --grpc-gateway_out=%WORK_DIR%\api\gen\v1 --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=%WORK_DIR%\api\gen\v1 --openapiv2_opt logtostderr=true ./serverB.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=paths=source_relative:%WORK_DIR%\api\gen\v1 --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./serverB_task.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=paths=source_relative:%WORK_DIR%\api\gen\v1 --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./serverB_CallBack.proto
@rem protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --go-asynq_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./server.proto
cd %WORK_DIR%
@rem pause