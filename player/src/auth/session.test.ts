import { extrairToken, urlLogin } from "./session";

describe("extrairToken", () => {
  it("extrai o token do fragmento", () => {
    expect(extrairToken("#token=abc.def.ghi")).toBe("abc.def.ghi");
  });
  it("extrai quando há outros parâmetros no fragmento", () => {
    expect(extrairToken("#state=x&token=abc%2Edef")).toBe("abc.def");
  });
  it("devolve null sem token", () => {
    expect(extrairToken("#nada=1")).toBeNull();
    expect(extrairToken("")).toBeNull();
  });
});

describe("urlLogin", () => {
  it("monta a URL de login com redirect_uri para /app", () => {
    expect(urlLogin("google", "https://gfcode.com.br")).toBe(
      "/auth/google?redirect_uri=https%3A%2F%2Fgfcode.com.br%2Fapp",
    );
  });
  it("suporta github", () => {
    expect(urlLogin("github", "https://gfcode.com.br")).toContain("/auth/github?redirect_uri=");
  });
});
