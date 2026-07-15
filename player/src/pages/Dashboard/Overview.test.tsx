import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Overview } from './Overview';

describe('Overview Dashboard Page', () => {
  it('renders the Hero Card (RF03)', () => {
    render(<Overview />);
    expect(screen.getByText('Build your Next Flow')).toBeTruthy();
    expect(screen.getByText('Get Started')).toBeTruthy();
  });

  it('renders the Status Cards (RF04)', () => {
    render(<Overview />);
    expect(screen.getByText('Active Projects')).toBeTruthy();
    expect(screen.getByText('Pending Tasks')).toBeTruthy();
    expect(screen.getByText('Team Members')).toBeTruthy();
  });

  it('renders the FAB (RF05)', () => {
    render(<Overview />);
    expect(screen.getByText('Create')).toBeTruthy();
  });
});
