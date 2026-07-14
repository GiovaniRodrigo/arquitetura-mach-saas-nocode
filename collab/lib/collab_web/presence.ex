defmodule CollabWeb.Presence do
  @moduledoc """
  Presença dos utilizadores num ecrã: quem está online e a posição do cursor,
  para colaboração em tempo real (RF06). Baseada em `Phoenix.Presence` (CRDT
  distribuído sobre o mesmo PubSub).
  """

  use Phoenix.Presence,
    otp_app: :collab,
    pubsub_server: Collab.PubSub
end
