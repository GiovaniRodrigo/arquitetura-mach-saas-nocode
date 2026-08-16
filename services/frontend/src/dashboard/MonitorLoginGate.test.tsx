import { describe, it, expect, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MonitorLoginGate } from './MonitorLoginGate';
import { sairDoMonitor, MONITOR_LOGIN, MONITOR_SENHA } from './monitorAuth';

afterEach(() => {
  sairDoMonitor();
});

function preencherELogar(login: string, senha: string) {
  fireEvent.change(screen.getByLabelText('Login'), { target: { value: login } });
  fireEvent.change(screen.getByLabelText('Senha'), { target: { value: senha } });
  fireEvent.click(screen.getByRole('button', { name: /entrar/i }));
}

describe('MonitorLoginGate', () => {
  it('não mostra o conteúdo antes do login', () => {
    render(
      <MonitorLoginGate>
        <p>Conteúdo protegido</p>
      </MonitorLoginGate>,
    );

    expect(screen.queryByText('Conteúdo protegido')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /entrar/i })).toBeInTheDocument();
  });

  it('login/senha incorretos mostram erro e mantêm o conteúdo escondido', () => {
    render(
      <MonitorLoginGate>
        <p>Conteúdo protegido</p>
      </MonitorLoginGate>,
    );

    preencherELogar('errado', 'errado');

    expect(screen.getByRole('alert')).toHaveTextContent(/inválidos/i);
    expect(screen.queryByText('Conteúdo protegido')).not.toBeInTheDocument();
  });

  it('login/senha corretos (definidos em monitorAuth.ts) liberam o conteúdo', () => {
    render(
      <MonitorLoginGate>
        <p>Conteúdo protegido</p>
      </MonitorLoginGate>,
    );

    preencherELogar(MONITOR_LOGIN, MONITOR_SENHA);

    expect(screen.getByText('Conteúdo protegido')).toBeInTheDocument();
  });

  it('mantém destravado numa nova montagem, dentro da mesma sessão do browser', () => {
    const { unmount } = render(
      <MonitorLoginGate>
        <p>Conteúdo protegido</p>
      </MonitorLoginGate>,
    );
    preencherELogar(MONITOR_LOGIN, MONITOR_SENHA);
    unmount();

    render(
      <MonitorLoginGate>
        <p>Conteúdo protegido</p>
      </MonitorLoginGate>,
    );

    expect(screen.getByText('Conteúdo protegido')).toBeInTheDocument();
  });
});
