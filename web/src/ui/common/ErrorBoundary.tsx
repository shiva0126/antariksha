import { Component, type ReactNode } from 'react';

/** Keeps one broken page from blanking the whole app: shows a short message
 *  with a reload button instead. Keyed by page, so navigating resets it. */
export class ErrorBoundary extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false };
  static getDerivedStateFromError() { return { failed: true }; }
  componentDidCatch(error: unknown) { console.error('page error', error); }
  render() {
    if (!this.state.failed) return this.props.children;
    return (
      <div className="page"><div className="card" role="alert">
        <h2>Something went wrong on this page</h2>
        <p className="muted">Your data is safe. Reloading usually fixes it.</p>
        <button className="primary" onClick={() => location.reload()}>Reload</button>
      </div></div>
    );
  }
}
