// Focus mode: no sidebar or tab bar while studying.
export default function FocusLayout({ children }: LayoutProps<"/[locale]">) {
  return <div className="min-h-[100dvh] bg-background">{children}</div>;
}
