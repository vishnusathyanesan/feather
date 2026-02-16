import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Highlight, themes } from "prism-react-renderer";

interface Props {
  content: string;
}

function highlightMentions(text: string): (string | JSX.Element)[] {
  const mentionRegex = /@(\w+)/g;
  const parts: (string | JSX.Element)[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = mentionRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }
    parts.push(
      <span
        key={match.index}
        className="rounded bg-blue-100 px-1 font-medium text-blue-800 dark:bg-blue-900/30 dark:text-blue-300"
      >
        {match[0]}
      </span>
    );
    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }

  return parts;
}

export default function MarkdownRenderer({ content }: Props) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || "");
          const code = String(children).replace(/\n$/, "");

          if (match) {
            return (
              <Highlight theme={themes.vsDark} code={code} language={match[1]}>
                {({ style, tokens, getLineProps, getTokenProps }) => (
                  <pre
                    className="my-2 overflow-x-auto rounded p-3 text-xs"
                    style={style}
                  >
                    {tokens.map((line, i) => (
                      <div key={i} {...getLineProps({ line })}>
                        {line.map((token, key) => (
                          <span key={key} {...getTokenProps({ token })} />
                        ))}
                      </div>
                    ))}
                  </pre>
                )}
              </Highlight>
            );
          }

          return (
            <code
              className="rounded bg-gray-100 px-1 py-0.5 text-xs dark:bg-gray-700"
              {...props}
            >
              {children}
            </code>
          );
        },
        a({ href, children }) {
          return (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:underline dark:text-blue-400"
            >
              {children}
            </a>
          );
        },
        p({ children }) {
          // Process text children to highlight @mentions
          const processed = Array.isArray(children)
            ? children.map((child, i) =>
                typeof child === "string" ? (
                  <span key={i}>{highlightMentions(child)}</span>
                ) : (
                  child
                )
              )
            : typeof children === "string"
            ? highlightMentions(children)
            : children;

          return <p className="mb-1 last:mb-0">{processed}</p>;
        },
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
