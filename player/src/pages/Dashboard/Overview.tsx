import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { FabButton } from '../../components/m3/FabButton';

export function Overview() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      {/* Hero Card */}
      <TonalCard className="bg-primary/10 text-primary border-none relative overflow-hidden">
        {/* Glow effect matching GF Code */}
        <div className="absolute -top-24 -right-24 w-64 h-64 bg-primary/20 blur-3xl rounded-full pointer-events-none" />
        <h2 className="text-2xl md:text-3xl font-heading font-bold mb-2">Build your Next Flow</h2>
        <p className="text-primary/80 mb-6 max-w-xl text-sm md:text-base font-medium">
          Start creating projects and designing your business architecture with our intuitive node-based editor.
        </p>
        <button className="bg-primary text-primary-foreground px-6 py-2.5 rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all shadow-sm">
          Get Started
        </button>
      </TonalCard>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <ElevatedCard className="bg-card text-card-foreground">
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Active Projects</h3>
          <p className="text-3xl font-heading font-bold text-foreground">12</p>
        </ElevatedCard>
        <ElevatedCard className="bg-card text-card-foreground">
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Pending Tasks</h3>
          <p className="text-3xl font-heading font-bold text-foreground">4</p>
        </ElevatedCard>
        <ElevatedCard className="bg-card text-card-foreground">
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Team Members</h3>
          <p className="text-3xl font-heading font-bold text-foreground">8</p>
        </ElevatedCard>
      </div>
      
      {/* FAB */}
      <FabButton icon="+" onClick={() => alert('Create new project')} className="md:bottom-8 md:right-8">
        Create
      </FabButton>
    </div>
  );
}
