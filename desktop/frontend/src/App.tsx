import { BarChart2, Clock, FolderOpen, Settings } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Shell } from "./components/layout/Shell";
import { Button } from "./components/ui/Button";
import { Modal } from "./components/ui/Modal";
import { OnboardingWizard, hasCompletedOnboarding } from "./components/ui/OnboardingWizard";
import { SplashScreen } from "./components/ui/SplashScreen";
import { ToastProvider } from "./components/ui/ToastProvider";
import { ViewTransition } from "./components/ui/ViewTransition";
import { useTheme } from "./hooks/useTheme";
import type { ViewRoute } from "./lib/types";
import { AnalyticsView } from "./views/AnalyticsView";
import { HistoryView } from "./views/HistoryView";
import { RenameView } from "./views/RenameView";
import { SettingsView } from "./views/SettingsView";
import type { NavItem } from "./lib/types";

const navItems: NavItem[] = [
    { route: "rename", label: "Rename", Icon: FolderOpen },
    { route: "history", label: "History", Icon: Clock },
    { route: "analytics", label: "Analytics", Icon: BarChart2 },
    { route: "settings", label: "Settings", Icon: Settings },
];

function App() {
    const [route, setRoute] = useState<ViewRoute>("rename");
    const { theme, setTheme, toggleTheme } = useTheme();
    const [dirtySettings, setDirtySettings] = useState(false);
    const [pendingRoute, setPendingRoute] = useState<ViewRoute | null>(null);
    const [splashDone, setSplashDone] = useState(false);
    const [onboardingDone, setOnboardingDone] = useState(() => hasCompletedOnboarding());
    const [stepIndex, setStepIndex] = useState(0);

    useEffect(() => {
      if (splashDone && !onboardingDone) {
        setOnboardingDone(hasCompletedOnboarding());
      }
    }, [splashDone, onboardingDone]);

    const handleRouteChange = useCallback(
        (nextRoute: ViewRoute) => {
            if (route === "settings" && dirtySettings && nextRoute !== "settings") {
                setPendingRoute(nextRoute);
                return;
            }
            setRoute(nextRoute);
        },
        [route, dirtySettings],
    );

    const handleDiscardAndNavigate = useCallback(() => {
        setDirtySettings(false);
        if (pendingRoute) {
            setRoute(pendingRoute);
            setPendingRoute(null);
        }
    }, [pendingRoute]);

    const handleDismissNavigate = useCallback(() => {
        setPendingRoute(null);
    }, []);

    return (
        <ToastProvider>
            {!splashDone ? <SplashScreen onDone={() => setSplashDone(true)} /> : null}
            {splashDone && !onboardingDone ? (
              <OnboardingWizard onDone={() => setOnboardingDone(true)} />
            ) : null}
            <Shell
                navItems={navItems}
                route={route}
                onRouteChange={handleRouteChange}
                theme={theme}
                onToggleTheme={toggleTheme}
                onOpenSettings={() => handleRouteChange("settings")}
                stepIndex={stepIndex}
            >
                {route === "rename" && <ViewTransition><RenameView onOpenSettings={() => handleRouteChange("settings")} onStepChange={setStepIndex} /></ViewTransition>}
                {route === "history" && <ViewTransition><HistoryView /></ViewTransition>}
                {route === "analytics" && <ViewTransition><AnalyticsView /></ViewTransition>}
                {route === "settings" && (
                    <ViewTransition>
                        <SettingsView
                            theme={theme}
                            onThemeChange={setTheme}
                            onDirtyChange={setDirtySettings}
                        />
                    </ViewTransition>
                )}
            </Shell>

            <Modal
                open={pendingRoute !== null}
                title="Unsaved Changes"
                onClose={handleDismissNavigate}
                footer={
                    <>
                        <Button variant="ghost" onClick={handleDismissNavigate}>
                            Cancel
                        </Button>
                        <Button variant="danger" onClick={handleDiscardAndNavigate}>
                            Discard
                        </Button>
                        <Button
                            variant="solid"
                            onClick={() => {
                                setPendingRoute(null);
                            }}
                        >
                            Go Back
                        </Button>
                    </>
                }
            >
                You have unsaved changes to your settings. Navigate away without saving?
            </Modal>
        </ToastProvider>
    );
}

export default App;
