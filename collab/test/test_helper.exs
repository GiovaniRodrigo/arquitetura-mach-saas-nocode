ExUnit.start()

# Mock do cliente gRPC do Design Engine (write-behind) — permite contar as chamadas
# de flush sem um servidor real.
Mox.defmock(Collab.Grpc.DesignClientMock, for: Collab.Grpc.DesignClient)
