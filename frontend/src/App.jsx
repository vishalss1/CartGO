import { useState } from "react";

const navItems = [
  { label: "OUR WORK", badge: "23" },
  { label: "ABOUT" },
  { label: "THE LATEST" },
  { label: "CAREERS" },
  { label: "CONTACT" },
];

function ArrowUpRightIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-3.5 w-3.5"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="2"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M7 17 17 7M9 7h8v8" />
    </svg>
  );
}

function ArrowDownIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-9 w-9"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="1.8"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m0 0-6-6m6 6 6-6" />
    </svg>
  );
}

function DotBurstIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="2"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v18M3 12h18M5.6 5.6l12.8 12.8M18.4 5.6 5.6 18.4" />
    </svg>
  );
}

function MenuIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-5 w-5"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="1.8"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 7h16M4 12h16M4 17h16" />
    </svg>
  );
}

function Navbar() {
  const [open, setOpen] = useState(false);

  return (
    <header className="border-b border-line">
      <nav className="flex min-h-[70px] items-center justify-between px-5 sm:px-8 lg:px-12 xl:px-16">
        <a
          href="#top"
          className="shrink-0 border-r border-line pr-6 text-[11px] font-semibold tracking-[0.42em] text-paper sm:pr-10"
        >
          CARTGO
        </a>

        <div className="hidden min-w-0 flex-1 items-center justify-center lg:flex">
          <ul className="flex items-center gap-8 xl:gap-12">
            {navItems.map((item) => (
              <li key={item.label}>
                <a
                  href="#"
                  className="group inline-flex items-center gap-2 text-[11px] font-semibold tracking-[0.2em] text-paper transition-opacity duration-200 hover:opacity-65"
                >
                  <span className="relative">
                    {item.label}
                    <span className="absolute -bottom-1 left-0 h-px w-0 bg-paper transition-all duration-200 group-hover:w-full" />
                  </span>
                  {item.badge ? (
                    <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-paper px-1.5 text-[10px] tracking-normal">
                      {item.badge}
                    </span>
                  ) : null}
                </a>
              </li>
            ))}
          </ul>
        </div>

        <div className="hidden border-l border-line pl-6 lg:block">
          <a
            href="#"
            className="inline-flex items-center gap-2 bg-paper px-4 py-3 text-sm font-medium text-surface transition-colors duration-200 hover:bg-white"
          >
            Get in touch
            <span className="text-accent">
              <DotBurstIcon />
            </span>
          </a>
        </div>

        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          className="inline-flex items-center justify-center rounded-full border border-line p-3 text-paper transition-colors duration-200 hover:border-paper lg:hidden"
          aria-label="Toggle menu"
          aria-expanded={open}
        >
          <MenuIcon />
        </button>
      </nav>

      {open ? (
        <div className="border-t border-line px-5 py-5 sm:px-8 lg:hidden">
          <div className="flex flex-col gap-4">
            {navItems.map((item) => (
              <a
                key={item.label}
                href="#"
                className="flex items-center justify-between text-sm font-semibold tracking-[0.18em] text-paper transition-opacity duration-200 hover:opacity-70"
              >
                <span>{item.label}</span>
                {item.badge ? (
                  <span className="inline-flex h-6 min-w-6 items-center justify-center rounded-full border border-paper px-1.5 text-[11px] tracking-normal">
                    {item.badge}
                  </span>
                ) : null}
              </a>
            ))}
            <a
              href="#"
              className="mt-2 inline-flex w-fit items-center gap-2 bg-paper px-4 py-3 text-sm font-medium text-surface transition-colors duration-200 hover:bg-white"
            >
              Get in touch
              <span className="text-accent">
                <DotBurstIcon />
              </span>
            </a>
          </div>
        </div>
      ) : null}
    </header>
  );
}

function Hero() {
  return (
    <section className="relative min-h-hero overflow-hidden">
      <div className="flex min-h-hero flex-col justify-between px-5 pb-8 pt-8 sm:px-8 sm:pb-10 sm:pt-10 lg:px-12 lg:pb-12 xl:px-16 xl:pb-14 xl:pt-12">
        <div className="w-full max-w-[1500px]">
          <h1 className="max-w-[14ch] text-left text-[4.35rem] font-black uppercase leading-[0.88] tracking-hero text-paper sm:text-[5.9rem] md:text-[7.4rem] lg:max-w-none lg:text-[8.7rem] xl:text-[10.5rem] 2xl:text-[12rem]">
            <span className="block">We exist</span>
            <span className="block text-accent">To beat</span>
            <span className="block">Indifference</span>
          </h1>
        </div>

        <div className="mt-10 border-t border-line pt-8 sm:pt-10">
          <div className="grid gap-10 lg:grid-cols-[1.15fr_0.95fr_0.9fr] lg:gap-8">
            <div className="max-w-[18rem]">
              <p className="text-[2.25rem] font-extrabold uppercase leading-[0.92] tracking-hero text-paper sm:text-[2.75rem]">
                The variable is a growth company
              </p>
              <a
                href="#"
                className="mt-8 inline-flex items-center gap-3 bg-paper px-5 py-4 text-base font-medium text-surface transition-colors duration-200 hover:bg-white"
              >
                <span>Our reel</span>
                <span className="text-accent">
                  <ArrowUpRightIcon />
                </span>
              </a>
            </div>

            <div className="flex max-w-[19rem] flex-col justify-between gap-8">
              <p className="max-w-[16rem] text-base font-medium leading-[1.25] text-muted sm:text-[1.05rem]">
                We combine innovation and advertising to accelerate business transformation.
              </p>
              <div className="flex flex-wrap gap-6 text-sm font-semibold text-paper">
                <a href="#" className="inline-flex items-center gap-2 transition-opacity duration-200 hover:opacity-65">
                  Instagram
                  <span className="inline-flex h-5 w-5 items-center justify-center rounded-full bg-paper text-surface">
                    <ArrowUpRightIcon />
                  </span>
                </a>
                <a href="#" className="inline-flex items-center gap-2 transition-opacity duration-200 hover:opacity-65">
                  LinkedIn
                  <span className="inline-flex h-5 w-5 items-center justify-center rounded-full bg-paper text-surface">
                    <ArrowUpRightIcon />
                  </span>
                </a>
              </div>
            </div>

            <div className="flex flex-col justify-between gap-10 text-sm text-muted lg:items-start lg:pt-1">
              <div className="space-y-1 leading-relaxed">
                <p>Based in</p>
                <p>HQ: Winston</p>
                <p>Salem, NC</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="pointer-events-none absolute bottom-8 right-6 hidden h-36 w-36 items-center justify-center rounded-full border border-line text-paper md:flex lg:bottom-10 lg:right-10 lg:h-44 lg:w-44 xl:h-48 xl:w-48">
        <ArrowDownIcon />
      </div>

      <div className="absolute bottom-[16.5rem] right-4 rotate-[-14deg] text-right sm:right-8 lg:right-16">
        <div className="inline-flex flex-col items-center gap-2">
          <span className="text-xs font-black uppercase leading-none tracking-[0.1em] text-paper">
            Give a
            <br />
            shift
          </span>
          <span className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-line bg-[#8B7CF0] text-sm font-black text-surface">
            GO
          </span>
        </div>
      </div>
    </section>
  );
}

export default function App() {
  return (
    <div id="top" className="min-h-screen bg-surface font-sans text-paper antialiased">
      <Navbar />
      <main>
        <Hero />
      </main>
    </div>
  );
}
