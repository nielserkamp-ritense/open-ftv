import { themeQuartz } from "ag-grid-community";
import type { RowSelectionOptions } from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import type { ComponentProps } from "react";

// Expose all AgGridReact props while providing sensible defaults that can be overridden
type GridProps<TData> = ComponentProps<typeof AgGridReact<TData>>;

export default function CondensedGrid<TData>(props: GridProps<TData>) {
    // Default theme for the grid; can be overridden via the `theme` prop
    const themeConfig = themeQuartz.withParams({
        fontFamily: "var(--font-sans)",
        fontSize: 14,
        borderColor: "#DEE2E6",
        borderRadius: 2,
        headerFontSize: 14,
        headerFontWeight: 700,
        headerRowBorder: {color: "#007BC7", width: 1},
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

    const { theme = themeConfig, rowSelection = defaultRowSelection, defaultColDef = defaultDef, ...rest } = props;

    return <AgGridReact theme={theme} rowSelection={rowSelection} defaultColDef={defaultDef} headerHeight={28} {...rest} />;
}