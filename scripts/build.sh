rm -rf internal/ui/dist/*
(cd dashboard && npm run build)
cp -r dashboard/out/* internal/ui/dist/
mkdir -p bin
go build -trimpath -o bin/tarak ./cmd/tarak
go build -trimpath -o bin/tarakctl ./cmd/tarakctl
go build -trimpath -o bin/taraktl ./cmd/taraktl
go build -trimpath -o bin/tarakd ./cmd/tarakd
go build -trimpath -o bin/taraks ./cmd/taraks
