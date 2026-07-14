defmodule Collab.Telemetry do
  @moduledoc """
  Instrumentação OpenTelemetry do motor de colaboração (RNF04): spans para os
  eventos de channel e para o flush write-behind, e propagação do `traceparent`
  para o Design Engine (Go), mantendo um único trace ponta-a-ponta entre Elixir e
  os microsserviços.

  Os atributos usam sempre identificadores anonimizados (tenant_id, blind_index),
  nunca nomes reais de campos (RNF08).
  """

  require OpenTelemetry.Tracer, as: Tracer

  @doc """
  Executa `fun` dentro de um span com o `name` e os `attributes` dados, devolvendo
  o valor de `fun`. Sem exportador (teste), o span é no-op.
  """
  @spec with_span(String.t(), map(), (-> result)) :: result when result: any()
  def with_span(name, attributes, fun) when is_function(fun, 0) do
    Tracer.with_span name, %{attributes: attributes} do
      fun.()
    end
  end

  @doc """
  Injeta o contexto de trace atual como cabeçalhos (`traceparent`/`tracestate`)
  sobre o mapa de metadata gRPC, para o serviço a jusante continuar o trace.
  """
  @spec inject_metadata(map()) :: map()
  def inject_metadata(metadata \\ %{}) do
    :otel_propagator_text_map.inject([])
    |> Enum.reduce(metadata, fn {k, v}, acc -> Map.put(acc, to_string(k), to_string(v)) end)
  end
end
