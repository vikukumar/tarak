import React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import { Mermaid } from "./Mermaid";

interface MarkdownViewerProps {
  content: string;
}

export const MarkdownViewer: React.FC<MarkdownViewerProps> = ({ content }) => {
  return (
    <div className="prose prose-invert max-w-none text-slate-300 text-sm leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
        components={{
          code({ node, inline, className, children, ...props }: any) {
            const match = /language-(\w+)/.exec(className || "");
            const language = match ? match[1] : "";
            const value = String(children).replace(/\n$/, "");

            if (!inline && language === "mermaid") {
              return <Mermaid chart={value} />;
            }

            if (!inline) {
              return (
                <div className="my-3 rounded-xl bg-[#04060c] border border-white/10 p-4 font-mono text-xs text-cyan-300 overflow-x-auto shadow-xl">
                  <code>{value}</code>
                </div>
              );
            }

            return (
              <code className="px-1.5 py-0.5 rounded bg-white/5 border border-white/10 text-cyan-300 font-mono text-xs" {...props}>
                {children}
              </code>
            );
          },
          table({ children }: any) {
            return (
              <div className="my-4 overflow-x-auto rounded-xl border border-white/10">
                <table className="w-full text-left text-xs border-collapse">
                  {children}
                </table>
              </div>
            );
          },
          th({ children }: any) {
            return (
              <th className="bg-slate-900/90 text-slate-200 font-bold p-3 border-b border-white/10 uppercase tracking-wider">
                {children}
              </th>
            );
          },
          td({ children }: any) {
            return <td className="p-3 border-b border-white/5 text-slate-300">{children}</td>;
          },
          h1({ children }: any) {
            return <h1 className="text-2xl font-bold text-white mt-6 mb-3">{children}</h1>;
          },
          h2({ children }: any) {
            return <h2 className="text-xl font-bold text-white mt-5 mb-2.5 text-cyan-300">{children}</h2>;
          },
          h3({ children }: any) {
            return <h3 className="text-base font-bold text-white mt-4 mb-2">{children}</h3>;
          },
          ul({ children }: any) {
            return <ul className="list-disc ml-5 my-2 space-y-1">{children}</ul>;
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
