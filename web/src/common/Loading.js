import React from "react";

const AI_COLORS = ["#404040", "#404040", "#404040"];

export const AiDots = ({size = "medium"}) => {
  const isSmall = size === "small";
  const isLarge = size === "large";
  const dotPx = isSmall ? 5 : isLarge ? 12 : 9;
  const gapPx = isSmall ? 5 : isLarge ? 10 : 8;
  const glowBlur = isSmall ? 0 : isLarge ? 12 : 8;

  return (
    <span style={{display: "inline-flex", alignItems: "center", gap: gapPx}}>
      {AI_COLORS.map((color, i) => (
        <span
          key={i}
          style={{
            display: "inline-block",
            width: dotPx,
            height: dotPx,
            borderRadius: "50%",
            background: color,
            animation: "casdoor-ai-bounce 1.4s ease-in-out infinite",
            animationDelay: `${i * 0.16}s`,
            boxShadow: glowBlur > 0 ? `0 0 ${glowBlur}px ${color}90` : "none",
          }}
        />
      ))}
    </span>
  );
};

const Loading = ({spinning = true, tip, type = "section", style}) => {
  if (!spinning) {return null;}

  const isPage = type === "page";
  const isSmall = type === "small";

  const wrapperStyle = {
    display: "flex",
    flexDirection: "column",
    justifyContent: "center",
    alignItems: "center",
    ...(isPage && {width: "100%", height: "calc(100vh - 120px)"}),
    ...(type === "section" && {padding: "48px 0"}),
    ...style,
  };

  return (
    <div style={wrapperStyle}>
      <AiDots size={isSmall ? "small" : isPage ? "large" : "medium"} />
      {tip && !isSmall && (
        <div style={{marginTop: 14, fontSize: 13, color: "#94A3B8", letterSpacing: "0.05em", fontWeight: 400}}>
          {tip}
        </div>
      )}
    </div>
  );
};

export default Loading;
