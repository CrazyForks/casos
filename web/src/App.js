import React, {Component} from "react";
import {withRouter} from "react-router-dom";
import {StyleProvider, legacyLogicalPropertiesTransformer} from "@ant-design/cssinjs";
import {ConfigProvider, FloatButton, Layout} from "antd";
import * as Setting from "./Setting";
import {getShadcnThemeComponents, getShadcnThemeToken} from "./shadcnTheme";
import ManagementPage from "./ManagementPage";

class App extends Component {
  constructor(props) {
    super(props);
    Setting.initServerUrl();
    this.state = {
      selectedMenuKey: 0,
      uri: null,
    };
  }

  componentDidUpdate() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      // eslint-disable-next-line no-restricted-globals
      this.setState({uri: location.pathname});
    }
  }

  renderContent() {
    return (
      <Layout id="parent-area">
        <ManagementPage
          uri={this.state.uri}
          history={this.props.history}
        />
      </Layout>
    );
  }

  render() {
    return (
      <React.Fragment>
        <ConfigProvider
          theme={{
            token: getShadcnThemeToken(),
            components: getShadcnThemeComponents(),
          }}>
          <StyleProvider hashPriority="high" transformers={[legacyLogicalPropertiesTransformer]}>
            <React.Fragment>
              <FloatButton.BackTop />
              {this.renderContent()}
            </React.Fragment>
          </StyleProvider>
        </ConfigProvider>
      </React.Fragment>
    );
  }
}

export default withRouter(App);
