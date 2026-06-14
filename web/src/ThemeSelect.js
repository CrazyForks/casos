import React from "react";
import {MoonOutlined, SunOutlined} from "@ant-design/icons";

class ThemeSelect extends React.Component {
  handleToggle = () => {
    const isDark = this.props.themeAlgorithm.includes("dark");
    this.props.onChange(isDark ? ["default"] : ["dark"]);
  };

  render() {
    const isDark = this.props.themeAlgorithm.includes("dark");
    const icon = isDark
      ? <SunOutlined style={{fontSize: "18px"}} />
      : <MoonOutlined style={{fontSize: "18px"}} />;

    return (
      <div className="select-box" onClick={this.handleToggle} style={{cursor: "pointer"}}>
        {icon}
      </div>
    );
  }
}

export default ThemeSelect;
