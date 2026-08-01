import { CopyButton } from "@/components/ui/copy-button";
import { cn } from "@/lib/utils";

interface CodeBlockProps {
  code: string;
  /** Shown on the right of the chrome bar, e.g. "bash" or "yaml". */
  lang?: string;
  /** Hide the traffic-light dots for a more compact inline block. */
  compact?: boolean;
  showLineNumbers?: boolean;
  className?: string;
}

export function CodeBlock({
  code,
  lang = "bash",
  compact = false,
  showLineNumbers = true,
  className,
}: CodeBlockProps) {
  const lines = code.split("\n");

  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-border bg-code-bg",
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-border px-4 py-2.5">
        {compact ? (
          <span className="font-mono text-xs text-muted-foreground">{lang}</span>
        ) : (
          <>
            <div className="flex items-center gap-1.5">
              <span className="size-3 rounded-full bg-border" />
              <span className="size-3 rounded-full bg-border" />
              <span className="size-3 rounded-full bg-border" />
            </div>
            <span className="font-mono text-xs text-muted-foreground">
              {lang}
            </span>
          </>
        )}
        <CopyButton text={code} />
      </div>

      <pre className="overflow-x-auto p-5 font-mono text-sm leading-relaxed">
        <code>
          {lines.map((line, i) => (
            <div key={i}>
              {showLineNumbers && (
                <span className="select-none pr-4 text-muted-foreground/40">
                  {String(i + 1).padStart(2, " ")}
                </span>
              )}
              {/* Comment lines get dimmed; everything else stays plain. */}
              <span
                className={
                  line.trimStart().startsWith("#")
                    ? "text-muted-foreground"
                    : "text-foreground"
                }
              >
                {line || " "}
              </span>
            </div>
          ))}
        </code>
      </pre>
    </div>
  );
}
