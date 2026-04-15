export default function SystemLogViewer({ logs }) {
  if (!logs || logs.length === 0) return null;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between border-b border-line pb-4">
        <h3 className="text-[11px] font-black uppercase tracking-[0.32em] text-muted">System Audit Logs</h3>
        <span className="text-[9px] font-bold uppercase tracking-widest text-accent opacity-60">Deterministic Trace Enabled</span>
      </div>
      
      <div className="space-y-4">
        {[...logs].reverse().map((log, index) => (
          <div key={index} className="group relative flex gap-6 overflow-hidden border border-line bg-surface/30 p-4 transition-all hover:border-paper/40">
            <div className="flex flex-col items-center">
              <div className="h-2 w-2 rounded-full bg-paper shadow-[0_0_8px_rgba(255,255,255,0.4)]" />
              <div className="h-full w-[1px] bg-line group-last:bg-transparent" />
            </div>
            
            <div className="flex-1 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[10px] font-bold uppercase tracking-widest text-paper">
                   {log.to.replaceAll("_", " ")}
                </span>
                <span className="text-[9px] font-medium text-muted">
                   {new Date(log.timestamp).toLocaleTimeString()}
                </span>
              </div>
              
              <div className="flex items-center gap-3">
                 <span className="bg-line px-2 py-0.5 text-[8px] font-black uppercase tracking-tighter text-muted">
                    {log.service}
                 </span>
                 <p className="text-[10px] italic text-muted opacity-70">
                    Transition from <span className="uppercase">{log.from || "NULL"}</span>
                 </p>
              </div>
            </div>

            {/* Micro-Metadata decoration */}
            <div className="absolute -right-4 -top-2 opacity-5 transition-opacity group-hover:opacity-10">
               <span className="text-[4rem] font-black uppercase italic leading-none">{log.service.split(" ")[0]}</span>
            </div>
          </div>
        ))}
      </div>

      <div className="flex items-center justify-center gap-2 pt-4">
         <div className="h-1 w-1 animate-ping rounded-full bg-accent" />
         <p className="text-[8px] font-bold uppercase tracking-[0.3em] text-muted opacity-50">
            Awaiting next state machine transition...
         </p>
      </div>
    </div>
  );
}
