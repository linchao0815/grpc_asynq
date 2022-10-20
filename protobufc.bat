set WORK_DIR=%cd%
set WORK_PROTO=%WORK_DIR%\api\v1
set INCLUDE_PROTO=%WORK_DIR%\api\gen
@rem del %WORK_DIR%\api\gen\v1 /S /Q
cd api/v1
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=%WORK_DIR%\api\gen\v1 --go_opt=paths=source_relative --go-grpc_out=%WORK_DIR%\api\gen\v1 --go-grpc_opt=paths=source_relative --grpc-gateway_out=%WORK_DIR%\api\gen\v1 --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=%WORK_DIR%\api\gen\v1 --openapiv2_opt logtostderr=true ./after.proto
@rem protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=paths=source_relative:%WORK_DIR%\api\gen\v1 --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./user.proto
protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=%WORK_DIR%\api\gen\v1 --go_opt=paths=source_relative --go-grpc_out=%WORK_DIR%\api\gen\v1 --go-grpc_opt=paths=source_relative --grpc-gateway_out=%WORK_DIR%\api\gen\v1 --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=%WORK_DIR%\api\gen\v1 --openapiv2_opt logtostderr=true --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./user.proto
@rem protoc -I=. --proto_path=%INCLUDE_PROTO% --proto_path=%GOPATH%\pkg\mod --zap-marshaler_out=%WORK_DIR% --go_out=paths=source_relative:%WORK_DIR%\api\gen\v1 --go-grpc_out=%WORK_DIR%\api\gen\v1 --go-grpc_opt=paths=source_relative --go-asynqgen_out=paths=source_relative:%WORK_DIR%\api\gen\v1 ./user.proto
cd %WORK_DIR%
@rem pause