import ReactMarkdown from 'react-markdown';
import type { Components } from 'react-markdown';

const components: Components = {
  p: ({ children }) => (
    <p style={{ margin: '0 0 0.5em', fontFamily: 'var(--font-serif)', color: 'var(--text-secondary)', fontSize: 'inherit', lineHeight: '1.6' }}>
      {children}
    </p>
  ),
  h1: ({ children }) => <h1 style={{ fontFamily: 'var(--font-serif)', color: 'var(--accent-primary)', fontSize: '1.4em', marginBottom: '0.5rem' }}>{children}</h1>,
  h2: ({ children }) => <h2 style={{ fontFamily: 'var(--font-serif)', color: 'var(--accent-primary)', fontSize: '1.2em', marginBottom: '0.4rem' }}>{children}</h2>,
  h3: ({ children }) => <h3 style={{ fontFamily: 'var(--font-serif)', color: 'var(--accent-primary)', fontSize: '1.05em', marginBottom: '0.3rem' }}>{children}</h3>,
  strong: ({ children }) => <strong style={{ fontWeight: 700, color: 'var(--text-primary)' }}>{children}</strong>,
  em: ({ children }) => <em style={{ fontStyle: 'italic' }}>{children}</em>,
  code: ({ children, className }) => {
    const isBlock = className?.startsWith('language-');
    return isBlock
      ? <code style={{ display: 'block', fontFamily: 'var(--font-mono)', fontSize: '0.85em', backgroundColor: 'var(--bg-primary)', padding: '0.5rem', overflowX: 'auto', color: 'var(--text-primary)' }}>{children}</code>
      : <code style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85em', backgroundColor: 'var(--bg-primary)', padding: '0.1em 0.3em', color: 'var(--accent-primary)' }}>{children}</code>;
  },
  pre: ({ children }) => <pre style={{ margin: '0.5em 0', backgroundColor: 'var(--bg-primary)', padding: '0.5rem', overflowX: 'auto' }}>{children}</pre>,
  ul: ({ children }) => <ul style={{ paddingLeft: '1.2em', margin: '0.3em 0', fontFamily: 'var(--font-serif)' }}>{children}</ul>,
  ol: ({ children }) => <ol style={{ paddingLeft: '1.2em', margin: '0.3em 0', fontFamily: 'var(--font-serif)' }}>{children}</ol>,
  li: ({ children }) => <li style={{ marginBottom: '0.15em', color: 'var(--text-secondary)' }}>{children}</li>,
  a: ({ href, children }) => <a href={href} target="_blank" rel="noreferrer" style={{ color: 'var(--accent-primary)', textDecoration: 'underline' }}>{children}</a>,
  blockquote: ({ children }) => (
    <blockquote style={{ borderLeft: '3px solid var(--accent-secondary)', paddingLeft: '0.75rem', margin: '0.5em 0', color: 'var(--text-muted)', fontStyle: 'italic', fontFamily: 'var(--font-serif)' }}>
      {children}
    </blockquote>
  ),
};

interface MarkdownRendererProps {
  content: string;
  fontSize?: string;
}

export function MarkdownRenderer({ content, fontSize }: MarkdownRendererProps) {
  return (
    <div style={{ fontSize: fontSize ?? 'inherit', lineHeight: '1.6', wordBreak: 'break-word' }}>
      <ReactMarkdown components={components}>{content}</ReactMarkdown>
    </div>
  );
}
