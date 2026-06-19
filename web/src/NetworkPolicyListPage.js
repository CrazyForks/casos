import React from "react";
import {
  Alert, Button, Checkbox, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag
} from "antd";
import {DeleteOutlined, EditOutlined, MinusCircleOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import * as NetworkPolicyBackend from "./backend/NetworkPolicyBackend";
import * as NamespaceBackend from "./backend/NamespaceBackend";
import * as Setting from "./Setting";

const POLICY_TYPES = ["Ingress", "Egress"];

class NetworkPolicyListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      networkpolicies: [],
      namespaces: [],
      loading: true,
      error: null,
      modalVisible: false,
      modalMode: "add",
      submitting: false,
      editingNp: null,
    };
    this.formRef = React.createRef();
  }

  componentDidMount() {
    this.fetchNetworkPolicies();
    this.fetchNamespaces();
  }

  fetchNamespaces() {
    NamespaceBackend.getNamespaces().then(res => {
      if (res.status === "ok") {
        this.setState({namespaces: res.data ?? []});
      }
    }).catch(() => {});
  }

  fetchNetworkPolicies() {
    this.setState({loading: true, error: null});
    NetworkPolicyBackend.getNetworkPolicies().then(res => {
      if (res.status === "ok") {
        this.setState({networkpolicies: res.data ?? []});
      } else {
        Setting.showMessage("error", res.msg);
        this.setState({error: res.msg});
      }
    }).catch(e => {
      Setting.showMessage("error", e.message);
      this.setState({error: e.message});
    }).finally(() => {
      this.setState({loading: false});
    });
  }

  openAddModal() {
    const defaultNs = this.state.namespaces.length > 0 ? this.state.namespaces[0].name : "default";
    this.setState({modalVisible: true, modalMode: "add", editingNp: null}, () => {
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({
          name: "",
          namespace: defaultNs,
          policyTypes: ["Ingress"],
          podSelectorEntries: [],
        });
      }, 0);
    });
  }

  openEditModal(np) {
    let podSelectorEntries = [];
    try {
      const labels = JSON.parse(np.podSelector || "{}");
      podSelectorEntries = Object.entries(labels).map(([key, value]) => ({key, value}));
    } catch {
      podSelectorEntries = [];
    }
    this.setState({modalVisible: true, modalMode: "edit", editingNp: np}, () => {
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({
          name: np.name,
          namespace: np.namespace,
          policyTypes: np.policyTypes ?? ["Ingress"],
          podSelectorEntries,
        });
      }, 0);
    });
  }

  closeModal() {
    this.setState({modalVisible: false, editingNp: null});
  }

  handleSubmit() {
    this.formRef.current?.validateFields().then(values => {
      const podSelectorLabels = {};
      (values.podSelectorEntries ?? []).forEach(({key, value}) => {
        if (key) {podSelectorLabels[key] = value ?? "";}
      });
      const payload = {
        name: values.name,
        namespace: values.namespace,
        policyTypes: values.policyTypes ?? [],
        podSelectorLabels,
        ingress: [],
        egress: [],
      };

      this.setState({submitting: true});

      if (this.state.modalMode === "add") {
        NetworkPolicyBackend.addNetworkPolicy(payload).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Network Policy created");
            this.closeModal();
            this.fetchNetworkPolicies();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      } else {
        const np = this.state.editingNp;
        NetworkPolicyBackend.updateNetworkPolicy({...payload, resourceVersion: np.resourceVersion}).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Network Policy updated");
            this.closeModal();
            this.fetchNetworkPolicies();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      }
    });
  }

  handleDelete(np) {
    NetworkPolicyBackend.deleteNetworkPolicy(np.namespace, np.name).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", "Network Policy deleted");
        this.fetchNetworkPolicies();
      } else {
        Setting.showMessage("error", res.msg);
      }
    }).catch(e => Setting.showMessage("error", e.message));
  }

  render() {
    const {networkpolicies, namespaces, loading, error, modalVisible, modalMode, submitting} = this.state;

    const nsOptions = namespaces.map(ns => ({label: ns.name, value: ns.name}));

    const columns = [
      {title: "Namespace", dataIndex: "namespace", key: "namespace", width: 160},
      {title: "Name", dataIndex: "name", key: "name"},
      {
        title: "Pod Selector",
        dataIndex: "podSelector",
        key: "podSelector",
        render: v => {
          try {
            const labels = JSON.parse(v || "{}");
            const entries = Object.entries(labels);
            if (entries.length === 0) {return <Tag>All Pods</Tag>;}
            return entries.map(([k, val]) => <Tag key={k}>{k}: {val}</Tag>);
          } catch {
            return <Tag>{v}</Tag>;
          }
        },
      },
      {
        title: "Policy Types",
        dataIndex: "policyTypes",
        key: "policyTypes",
        width: 160,
        render: types => (types ?? []).map(t => <Tag key={t} color={t === "Ingress" ? "blue" : "orange"}>{t}</Tag>),
      },
      {title: "Ingress Rules", dataIndex: "ingressRules", key: "ingressRules", width: 120},
      {title: "Egress Rules", dataIndex: "egressRules", key: "egressRules", width: 120},
      {title: "Created", dataIndex: "createdAt", key: "createdAt", width: 180},
      {
        title: "Actions",
        key: "actions",
        width: 140,
        render: (_, record) => (
          <Space>
            <Button size="small" icon={<EditOutlined />} onClick={() => this.openEditModal(record)}>Edit</Button>
            <Popconfirm
              title={`Delete Network Policy "${record.name}"?`}
              okText="Delete"
              okType="danger"
              cancelText="Cancel"
              onConfirm={() => this.handleDelete(record)}
            >
              <Button size="small" danger icon={<DeleteOutlined />}>Delete</Button>
            </Popconfirm>
          </Space>
        ),
      },
    ];

    return (
      <div style={{padding: "24px"}}>
        {error && (
          <Alert
            type="error"
            message="Failed to fetch Network Policies"
            description={error}
            style={{marginBottom: 16}}
            showIcon
          />
        )}

        <Table
          rowKey={r => `${r.namespace}/${r.name}`}
          columns={columns}
          dataSource={networkpolicies}
          loading={loading}
          size="middle"
          pagination={{pageSize: 20}}
          locale={{emptyText: "No Network Policies found"}}
          title={() => (
            <div>
              <span style={{fontWeight: 600}}>Network Policies</span>
              &nbsp;&nbsp;&nbsp;&nbsp;
              <Button icon={<ReloadOutlined />} onClick={() => this.fetchNetworkPolicies()} loading={loading} size="small">
                Refresh
              </Button>
              &nbsp;&nbsp;
              <Button type="primary" icon={<PlusOutlined />} size="small" onClick={() => this.openAddModal()}>
                Add
              </Button>
            </div>
          )}
        />

        <Modal
          title={modalMode === "add" ? "Add Network Policy" : "Edit Network Policy"}
          open={modalVisible}
          onOk={() => this.handleSubmit()}
          onCancel={() => this.closeModal()}
          confirmLoading={submitting}
          okText={modalMode === "add" ? "Create" : "Update"}
          width={580}
          destroyOnHidden
        >
          <Form ref={this.formRef} layout="vertical">
            <Form.Item label="Namespace" name="namespace" rules={[{required: true, message: "Namespace is required"}]}>
              <Select disabled={modalMode === "edit"} options={nsOptions} placeholder="Select a namespace" showSearch />
            </Form.Item>
            <Form.Item label="Name" name="name" rules={[{required: true, message: "Name is required"}]}>
              <Input disabled={modalMode === "edit"} placeholder="my-network-policy" />
            </Form.Item>
            <Form.Item label="Policy Types" name="policyTypes" rules={[{required: true, message: "At least one policy type is required"}]}>
              <Checkbox.Group options={POLICY_TYPES} />
            </Form.Item>

            <Form.List name="podSelectorEntries">
              {(fields, {add, remove}) => (
                <>
                  <div style={{marginBottom: 8, fontWeight: 500}}>
                    Pod Selector Labels
                    <span style={{fontWeight: 400, color: "#888", marginLeft: 8, fontSize: 12}}>
                      (leave empty to select all pods)
                    </span>
                  </div>
                  {fields.map(({key, name, ...rest}) => (
                    <Space key={key} align="baseline" style={{display: "flex", marginBottom: 4}}>
                      <Form.Item
                        {...rest}
                        name={[name, "key"]}
                        rules={[{required: true, message: "Key required"}]}
                        style={{marginBottom: 0}}
                      >
                        <Input placeholder="key" style={{width: 180}} />
                      </Form.Item>
                      <Form.Item
                        {...rest}
                        name={[name, "value"]}
                        style={{marginBottom: 0}}
                      >
                        <Input placeholder="value" style={{width: 220}} />
                      </Form.Item>
                      <MinusCircleOutlined onClick={() => remove(name)} style={{color: "#ff4d4f", cursor: "pointer"}} />
                    </Space>
                  ))}
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    icon={<PlusOutlined />}
                    style={{marginTop: 4}}
                    size="small"
                  >
                    Add Label
                  </Button>
                </>
              )}
            </Form.List>
          </Form>
        </Modal>
      </div>
    );
  }
}

export default NetworkPolicyListPage;
