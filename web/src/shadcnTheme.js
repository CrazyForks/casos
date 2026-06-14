const DefaultColorPrimary = "#404040";

const structuralTokens = {
  fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif",
  borderRadius: 10,
  borderRadiusXS: 2,
  borderRadiusSM: 6,
  borderRadiusLG: 14,
  padding: 16,
  paddingSM: 12,
  paddingLG: 24,
  margin: 16,
  marginSM: 12,
  marginLG: 24,
  boxShadow: "0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)",
  boxShadowSecondary: "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)",
};

const lightColorTokens = {
  colorPrimary: DefaultColorPrimary,
  colorSuccess: "#22c55e",
  colorWarning: "#f97316",
  colorError: "#ef4444",
  colorInfo: DefaultColorPrimary,
  colorTextBase: DefaultColorPrimary,
  colorBgBase: "#ffffff",
  colorPrimaryBg: "#f5f5f5",
  colorPrimaryBgHover: "#e5e5e5",
  colorPrimaryBorder: "#d4d4d4",
  colorPrimaryBorderHover: "#a3a3a3",
  colorPrimaryHover: "#404040",
  colorPrimaryActive: "#171717",
  colorPrimaryText: DefaultColorPrimary,
  colorPrimaryTextHover: "#404040",
  colorPrimaryTextActive: "#171717",
  colorLink: DefaultColorPrimary,
  colorLinkHover: "#404040",
  colorLinkActive: "#171717",
  colorText: DefaultColorPrimary,
  colorTextSecondary: "#525252",
  colorTextTertiary: "#737373",
  colorTextQuaternary: "#a3a3a3",
  colorTextDisabled: "#a3a3a3",
  colorBgContainer: "#ffffff",
  colorBgElevated: "#ffffff",
  colorBgLayout: "#fafafa",
  colorBorder: "#e5e5e5",
  colorBorderSecondary: "#f5f5f5",
};

const darkLinkTokens = {
  colorLink: "#d4d4d4",
  colorLinkHover: "#f5f5f5",
  colorLinkActive: "#a3a3a3",
};

export function getShadcnThemeToken(isDark) {
  if (isDark) {
    return {...structuralTokens, ...darkLinkTokens};
  }
  return {...structuralTokens, ...lightColorTokens};
}

function getLightComponents() {
  return {
    Button: {
      primaryShadow: "none",
      defaultShadow: "none",
      dangerShadow: "none",
      defaultBorderColor: "#e4e4e7",
      defaultColor: "#18181b",
      defaultBg: "#ffffff",
      defaultHoverBg: "#f4f4f5",
      defaultHoverBorderColor: "#d4d4d8",
      defaultHoverColor: "#18181b",
      defaultActiveBg: "#e4e4e7",
      defaultActiveBorderColor: "#d4d4d8",
      borderRadius: 6,
    },
    Input: {
      activeShadow: "none",
      hoverBorderColor: "#a1a1aa",
      activeBorderColor: "#71717a",
      borderRadius: 6,
    },
    Select: {
      optionSelectedBg: "#f4f4f5",
      optionActiveBg: "#fafafa",
      optionSelectedFontWeight: 500,
      borderRadius: 6,
    },
    Alert: {borderRadiusLG: 8},
    Modal: {borderRadiusLG: 12},
    Progress: {defaultColor: "#18181b", remainingColor: "#f4f4f5"},
    Steps: {iconSize: 32},
    Switch: {trackHeight: 24, trackMinWidth: 44, innerMinMargin: 4, innerMaxMargin: 24},
    Checkbox: {borderRadiusSM: 4},
    Menu: {
      itemFontSize: 14,
      groupTitleFontSize: 12,
      itemHeight: 40,
      fontWeightStrong: 600,
      itemSelectedBg: "rgba(0, 0, 0, 0.12)",
      itemSelectedColor: "inherit",
    },
    Table: {headerBg: "#fafafa", headerSplitColor: "#e5e5e5", fontWeightStrong: 600},
  };
}

function getDarkComponents() {
  return {
    Button: {primaryShadow: "none", defaultShadow: "none", dangerShadow: "none", borderRadius: 6},
    Input: {activeShadow: "none", hoverBorderColor: "#555", activeBorderColor: "#888", borderRadius: 6},
    Select: {
      optionSelectedFontWeight: 500,
      optionActiveBg: "rgba(255, 255, 255, 0.08)",
      optionSelectedBg: "rgba(255, 255, 255, 0.12)",
      borderRadius: 6,
    },
    Alert: {borderRadiusLG: 8},
    Modal: {borderRadiusLG: 12},
    Steps: {iconSize: 32},
    Switch: {trackHeight: 24, trackMinWidth: 44, innerMinMargin: 4, innerMaxMargin: 24},
    Checkbox: {borderRadiusSM: 4},
    Menu: {
      itemFontSize: 14,
      groupTitleFontSize: 12,
      itemHeight: 40,
      fontWeightStrong: 600,
      itemHoverBg: "rgba(255, 255, 255, 0.08)",
      itemSelectedBg: "rgba(255, 255, 255, 0.12)",
      itemSelectedColor: "inherit",
    },
    Table: {fontWeightStrong: 600},
  };
}

export function getShadcnThemeComponents(isDark) {
  return isDark ? getDarkComponents() : getLightComponents();
}
