defmodule Collab.TelemetryTest do
  @moduledoc "Instrumentação OTel: spans e propagação de traceparent (RNF04)."

  use ExUnit.Case, async: true

  test "with_span executa a função e devolve o seu valor" do
    assert Collab.Telemetry.with_span("t", %{"platform.tenant_id" => "t1"}, fn -> :ok end) == :ok
  end

  test "inject_metadata preserva a metadata existente" do
    md = Collab.Telemetry.inject_metadata(%{"x-tenant-context-bin" => "abc"})
    assert md["x-tenant-context-bin"] == "abc"
    assert is_map(md)
  end

  test "inject_metadata adiciona traceparent dentro de um span" do
    Collab.Telemetry.with_span("t", %{}, fn ->
      md = Collab.Telemetry.inject_metadata(%{})
      assert Map.has_key?(md, "traceparent"), "esperava traceparent no metadata dentro do span"
    end)
  end
end
