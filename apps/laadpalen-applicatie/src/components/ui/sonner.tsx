import { useTheme } from "next-themes";
import { AlertTriangle } from "lucide-react";
import { Toaster as Sonner, toast } from "sonner";

type ToasterProps = React.ComponentProps<typeof Sonner>;

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      icons={{ error: <AlertTriangle className="h-4 w-4"/> }}
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg",
          title: "group-data-[type=error]:text-red-700 dark:group-data-[type=error]:text-red-400",
          description: "group-[.toast]:text-muted-foreground",
          actionButton: "group-[.toast]:bg-primary group-[.toast]:text-primary-foreground",
          cancelButton: "group-[.toast]:bg-muted group-[.toast]:text-muted-foreground",
          icon:
            "group-data-[type=error]:text-red-600 dark:group-data-[type=error]:text-red-400 group-data-[type=error]:!h-8 group-data-[type=error]:!w-8 group-data-[type=error]:!bg-red-100 dark:group-data-[type=error]:!bg-red-950 group-data-[type=error]:!rounded-full group-data-[type=error]:!justify-center",
          // Matches the old destructive Toast variant: white/zinc bg (not a solid red block),
          // red border + left accent bar, red icon and title.
          error:
            "group-[.toaster]:!bg-white dark:group-[.toaster]:!bg-zinc-900 group-[.toaster]:!border-red-200 dark:group-[.toaster]:!border-red-900/50 group-[.toaster]:relative group-[.toaster]:after:content-[''] group-[.toaster]:after:absolute group-[.toaster]:after:top-0 group-[.toaster]:after:left-0 group-[.toaster]:after:h-full group-[.toaster]:after:w-1 group-[.toaster]:after:bg-red-500 group-[.toaster]:after:rounded-l-md",
        },
      }}
      {...props}
    />
  );
};

export { Toaster, toast };
