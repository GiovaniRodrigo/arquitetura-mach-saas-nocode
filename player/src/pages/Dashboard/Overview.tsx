import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { FabButton } from '../../components/m3/FabButton';

export function Overview() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      {/* Hero Card */}
      <TonalCard className="bg-blue-100 text-blue-900 border-none">
        <h2 className="text-2xl md:text-3xl font-semibold mb-2">Build your Next Flow</h2>
        <p className="text-blue-800 mb-6 max-w-xl">
          Start creating projects and designing your business architecture with our intuitive node-based editor.
        </p>
        <button className="bg-blue-600 text-white px-6 py-2.5 rounded-full font-medium hover:bg-blue-700 active:scale-95 transition-all shadow-sm">
          Get Started
        </button>
      </TonalCard>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <ElevatedCard>
          <h3 className="text-sm font-medium text-slate-500 mb-1">Active Projects</h3>
          <p className="text-3xl font-bold text-slate-800">12</p>
        </ElevatedCard>
        <ElevatedCard>
          <h3 className="text-sm font-medium text-slate-500 mb-1">Pending Tasks</h3>
          <p className="text-3xl font-bold text-slate-800">4</p>
        </ElevatedCard>
        <ElevatedCard>
          <h3 className="text-sm font-medium text-slate-500 mb-1">Team Members</h3>
          <p className="text-3xl font-bold text-slate-800">8</p>
        </ElevatedCard>
      </div>
      
      {/* FAB */}
      <FabButton icon="+" onClick={() => alert('Create new project')} className="md:bottom-8 md:right-8">
        Create
      </FabButton>
    </div>
  );
}
