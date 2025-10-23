import { Suspense, lazy } from "react";
import { themeQuartz } from "ag-grid-community";
import type { RowSelectionOptions } from "ag-grid-community";

// Lazily load the heavy ag-grid-react component to enable code-splitting
const LazyAgGridReact = lazy(async () => {
  const mod = await import("ag-grid-react");
  return { default: mod.AgGridReact };
});

// Keep prop types lightweight to avoid pulling ag-grid types into the main bundle
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type GridProps = { [key: string]: any };

export default function CondensedGrid(props: GridProps) {
  // Default theme for the grid; can be overridden via the `theme` prop
  const themeConfig = themeQuartz.withParams({
    fontFamily: "var(--font-sans)",
    fontSize: 14,
    borderColor: "#DEE2E6",
    borderRadius: 2,
    headerFontSize: 14,
    headerFontWeight: 700,
    headerRowBorder: { color: "#007BC7", width: 1 },
    headerColumnBorder: false,
    headerTextColor: "#007BC7",
    spacing: 8,
    wrapperBorder: false,
    wrapperBorderRadius: 4,
    headerBackgroundColor: "#FFFFFF",
    headerVerticalPaddingScale: 1.2,
  });

  const defaultRowSelection: RowSelectionOptions = {
    mode: "multiRow",
    checkboxes: true,
    enableClickSelection: true,
  };
  const defaultDef = {
    flex: 1,
  };

  const { theme = themeConfig, rowSelection = defaultRowSelection, defaultColDef = defaultDef, ...rest } = props as any;

  return (
    <Suspense fallback={<div className="p-2 text-xs text-content-secondary">Loading grid…</div>}>
      <LazyAgGridReact
        theme={theme}
        rowSelection={rowSelection}
        defaultColDef={defaultDef}
        headerHeight={28}
        {...rest}
      />
    </Suspense>
  );
}