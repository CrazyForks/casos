import React, {Component} from "react";
import {Redirect, Route, Switch, withRouter} from "react-router-dom";
import {StyleProvider, legacyLogicalPropertiesTransformer} from "@ant-design/cssinjs";
import {ConfigProvider, FloatButton, Layout} from "antd";
import * as Setting from "./Setting";
import * as AccountBackend from "./backend/AccountBackend";
import * as Conf from "./Conf";
import {getShadcnThemeComponents, getShadcnThemeToken} from "./shadcnTheme";
import ManagementPage from "./ManagementPage";
import AuthCallback from "./AuthCallback";
import SigninPage from "./SigninPage";

class App extends Component {
  constructor(props) {
    super(props);
    Setting.initServerUrl();
    Setting.initCasdoorSdk(Conf.AuthConfig);
    this.state = {
      selectedMenuKey: 0,
      uri: null,
      account: undefined,
    };
  }

  UNSAFE_componentWillMount() {
    this.getAccount();
  }

  componentDidUpdate() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      // eslint-disable-next-line no-restricted-globals
      this.setState({uri: location.pathname});
    }
  }

  getAccount() {
    AccountBackend.getAccount().then((res) => {
      const account = res.data;
      this.setState({account: account});
    });
  }

  signout() {
    AccountBackend.signout().then((res) => {
      if (res.status === "ok") {
        this.setState({account: null});
        Setting.showMessage("success", "Successfully signed out");
        Setting.goToLink("/");
      } else {
        Setting.showMessage("error", `Signout failed: ${res.msg}`);
      }
    });
  }

  renderHomeIfSignedIn(component) {
    if (this.state.account !== null && this.state.account !== undefined) {
      return <Redirect to="/" />;
    } else {
      return component;
    }
  }

  renderSigninIfNotSignedIn(component) {
    if (this.state.account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/signin" />;
    } else if (this.state.account === undefined) {
      return null;
    } else {
      return component;
    }
  }

  renderContent() {
    return (
      <Layout id="parent-area">
        <Switch>
          <Route exact path="/callback" component={AuthCallback} />
          <Route exact path="/signin" render={(props) => this.renderHomeIfSignedIn(<SigninPage {...props} />)} />
          <Route path="/" render={(props) => this.renderSigninIfNotSignedIn(
            <ManagementPage
              account={this.state.account}
              uri={this.state.uri}
              history={this.props.history}
              onSignout={this.signout.bind(this)}
              {...props}
            />
          )} />
        </Switch>
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
