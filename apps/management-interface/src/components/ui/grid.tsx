import { Suspense, lazy } from "react";
import { themeQuartz } from "ag-grid-community";
import type { RowSelectionOptions } from "ag-grid-community";

// Lazily load the heavy ag-grid-react component to enable code-splitting
const LazyAgGridReact = lazy(async () => {
  const mod = await import("ag-grid-react");
  return { default: mod.AgGridReact };
});

// Keep prop types lightweight to avoid pulling ag-grid types into the main bundle
// If you need strong typing, you can import types from 'ag-grid-react' as type-only imports.
// For bundle size, we keep it generic here.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type GridProps = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
};

export default function Grid(props: GridProps) {
  // Default theme for the grid; can be overridden via the `theme` prop
  const themeConfig = themeQuartz.withParams({
    fontFamily: "var(--font-sans)",
    fontSize: 16,
    borderColor: "#DEE2E6",
    borderRadius: 2,
    headerFontSize: 14,
    headerRowBorder: true,
    headerTextColor: "var(--color-content-secondary)",
    spacing: 8,
    wrapperBorder: false,
    wrapperBorderRadius: 4,
    headerBackgroundColor: "#FFFFFF",
    headerVerticalPaddingScale: 1.2,
  });

  const defaultRowSelection: RowSelectionOptions = {
    mode: "singleRow",
    checkboxes: false,
    enableClickSelection: true,
  };
  const defaultDef = {
    flex: 1,
  };

  const { theme = themeConfig, rowSelection = defaultRowSelection, defaultColDef = defaultDef, ...rest } = props as any;

  return (
    <Suspense fallback={<div className="p-4 text-sm text-content-secondary">Loading grid…</div>}>
      <LazyAgGridReact theme={theme} rowSelection={rowSelection} defaultColDef={defaultDef} {...rest} />
    </Suspense>
  );
}