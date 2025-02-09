@echo off
:: Set protoc and plugin paths
set PROTOC_GEN_GO=D:\GoWork\bin\protoc-gen-go.exe
set PROTOC_GEN_GO_GRPC=D:\GoWork\bin\protoc-gen-go-grpc.exe
set PROTOC=D:\protoc-29.3-win64\bin\protoc.exe
set PATH=%PATH%;D:\GoWork\bin;D:\protoc-29.3-win64\bin

:: Generate proto code for all services
"%PROTOC%" --plugin=protoc-gen-go=%PROTOC_GEN_GO% ^
       --plugin=protoc-gen-go-grpc=%PROTOC_GEN_GO_GRPC% ^
       --go_out=. --go_opt=paths=source_relative ^
       --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
       proto/user/user.proto proto/auth/auth.proto ^
       proto/product/product.proto proto/cart/cart.proto ^
       proto/order/order.proto proto/payment/payment.proto ^
       proto/checkout/checkout.proto proto/ai/ai.proto
