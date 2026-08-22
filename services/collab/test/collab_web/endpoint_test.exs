defmodule CollabWeb.EndpointTest do
  use CollabWeb.ConnCase, async: true

  # RF02/RN02 (specs/008-monitor-recursos): /healthz passa a responder um corpo
  # JSON com status, uptime e memória alocada da VM, mantendo o status HTTP 200
  # já usado pelo smoke test pós-deploy (spec 002, RF11).
  test "GET /healthz retorna 200 com corpo JSON de recursos", %{conn: conn} do
    conn = get(conn, "/healthz")

    assert conn.status == 200
    assert Plug.Conn.get_resp_header(conn, "content-type") |> List.first() =~ "application/json"

    body = Jason.decode!(conn.resp_body)

    assert body["status"] == "servindo"
    assert is_integer(body["uptime_segundos"])
    assert body["uptime_segundos"] >= 0
    assert is_integer(body["memoria_alocada_bytes"])
    assert body["memoria_alocada_bytes"] > 0

    # RN02/plan.md §4.3: BEAM não expõe equivalente a memória de sistema/goroutines,
    # então esses campos simplesmente não aparecem no corpo (não são inventados).
    refute Map.has_key?(body, "memoria_sistema_bytes")
    refute Map.has_key?(body, "goroutines")
  end
end
