import React from "react";
import {
  Alert, Button, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Tooltip
} from "antd";
import {DeleteOutlined, EditOutlined, MinusCircleOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import * as ResourceQuotaBackend from "./backend/ResourceQuotaBackend";
import * as NamespaceBackend from "./backend/NamespaceBackend";
import * as Setting from "./Setting";

const COMMON_RESOURCES = [
  "requests.cpu",
  "requests.memory",
  "limits.cpu",
  "limits.memory",
  "pods",
  "services",
  "configmaps",
  "secrets",
  "persistentvolumeclaims",
  "services.loadbalancers",
  "services.nodeports",
];

class ResourceQuotaListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      resourcequotas: [],
      namespaces: [],
      loading: true,
      error: null,
      modalVisible: false,
      modalMode: "add",
      submitting: false,
      editingRq: null,
    };
    this.formRef = React.createRef();
  }

  componentDidMount() {
    this.fetchResourceQuotas();
    this.fetchNamespaces();
  }

  fetchNamespaces() {
    NamespaceBackend.getNamespaces().then(res => {
      if (res.status === "ok") {
        this.setState({namespaces: res.data ?? []});
      }
    }).catch(() => {});
  }

  fetchResourceQuotas() {
    this.setState({loading: true, error: null});
    ResourceQuotaBackend.getResourceQuotas().then(res => {
      if (res.status === "ok") {
        this.setState({resourcequotas: res.data ?? []});
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
    this.setState({modalVisible: true, modalMode: "add", editingRq: null}, () => {
      const defaultNs = this.state.namespaces.length > 0 ? this.state.namespaces[0].name : "default";
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({name: "", namespace: defaultNs, hardEntries: []});
      }, 0);
    });
  }

  openEditModal(rq) {
    const hardEntries = Object.entries(rq.hard ?? {}).map(([resource, quantity]) => ({resource, quantity}));
    this.setState({modalVisible: true, modalMode: "edit", editingRq: rq}, () => {
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({
          name: rq.name,
          namespace: rq.namespace,
          hardEntries,
        });
      }, 0);
    });
  }

  closeModal() {
    this.setState({modalVisible: false, editingRq: null});
  }

  handleSubmit() {
    this.formRef.current?.validateFields().then(values => {
      const hard = {};
      (values.hardEntries ?? []).forEach(({resource, quantity}) => {
        if (resource && quantity) {
          hard[resource] = quantity;
        }
      });
      const payload = {
        name: values.name,
        namespace: values.namespace,
        hard,
      };

      this.setState({submitting: true});

      if (this.state.modalMode === "add") {
        ResourceQuotaBackend.addResourceQuota(payload).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Resource Quota created");
            this.closeModal();
            this.fetchResourceQuotas();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      } else {
        const rq = this.state.editingRq;
        ResourceQuotaBackend.updateResourceQuota({
          ...payload,
          resourceVersion: rq.resourceVersion,
        }).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Resource Quota updated");
            this.closeModal();
            this.fetchResourceQuotas();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      }
    });
  }

  handleDelete(rq) {
    ResourceQuotaBackend.deleteResourceQuota(rq.namespace, rq.name).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", "Resource Quota deleted");
        this.fetchResourceQuotas();
      } else {
        Setting.showMessage("error", res.msg);
      }
    }).catch(e => Setting.showMessage("error", e.message));
  }

  renderHardLimits(hard) {
    if (!hard || Object.keys(hard).length === 0) {
      return <span style={{color: "#aaa"}}>—</span>;
    }
    const entries = Object.entries(hard);
    const visible = entries.slice(0, 3);
    const rest = entries.slice(3);
    return (
      <span>
        {visible.map(([k, v]) => (
          <Tag key={k} style={{marginBottom: 2}}>{k}: {v}</Tag>
        ))}
        {rest.length > 0 && (
          <Tooltip title={rest.map(([k, v]) => `${k}: ${v}`).join("\n")}>
            <Tag>+{rest.length} more</Tag>
          </Tooltip>
        )}
      </span>
    );
  }

  render() {
    const {resourcequotas, namespaces, loading, error, modalVisible, modalMode, submitting} = this.state;

    const nsOptions = namespaces.map(ns => ({label: ns.name, value: ns.name}));
    const resourceOptions = COMMON_RESOURCES.map(r => ({label: r, value: r}));

    const columns = [
      {title: "Namespace", dataIndex: "namespace", key: "namespace", width: 160},
      {title: "Name", dataIndex: "name", key: "name", width: 200},
      {
        title: "Hard Limits",
        dataIndex: "hard",
        key: "hard",
        render: (hard) => this.renderHardLimits(hard),
      },
      {
        title: "Used",
        dataIndex: "used",
        key: "used",
        render: (used) => {
          if (!used || Object.keys(used).length === 0) {
            return <span style={{color: "#aaa"}}>—</span>;
          }
          const entries = Object.entries(used).slice(0, 3);
          return (
            <span>
              {entries.map(([k, v]) => (
                <Tag key={k} color="blue" style={{marginBottom: 2}}>{k}: {v}</Tag>
              ))}
            </span>
          );
        },
      },
      {title: "Created", dataIndex: "createdAt", key: "createdAt", width: 180},
      {
        title: "Actions",
        key: "actions",
        width: 140,
        render: (_, record) => (
          <Space>
            <Button size="small" icon={<EditOutlined />} onClick={() => this.openEditModal(record)}>Edit</Button>
            <Popconfirm
              title={`Delete Resource Quota "${record.name}"?`}
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
            message="Failed to fetch Resource Quotas"
            description={error}
            style={{marginBottom: 16}}
            showIcon
          />
        )}

        <Table
          rowKey={r => `${r.namespace}/${r.name}`}
          columns={columns}
          dataSource={resourcequotas}
          loading={loading}
          size="middle"
          scroll={{x: 900}}
          pagination={{pageSize: 20}}
          locale={{emptyText: "No Resource Quotas found"}}
          title={() => (
            <div>
              <span style={{fontWeight: 600}}>Resource Quotas</span>
              &nbsp;&nbsp;&nbsp;&nbsp;
              <Button icon={<ReloadOutlined />} onClick={() => this.fetchResourceQuotas()} loading={loading} size="small">
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
          title={modalMode === "add" ? "Add Resource Quota" : "Edit Resource Quota"}
          open={modalVisible}
          onOk={() => this.handleSubmit()}
          onCancel={() => this.closeModal()}
          confirmLoading={submitting}
          okText={modalMode === "add" ? "Create" : "Update"}
          width={620}
          destroyOnHidden
        >
          <Form ref={this.formRef} layout="vertical">
            <Form.Item
              label="Namespace"
              name="namespace"
              rules={[{required: true, message: "Namespace is required"}]}
            >
              <Select
                disabled={modalMode === "edit"}
                options={nsOptions}
                placeholder="Select a namespace"
                showSearch
              />
            </Form.Item>
            <Form.Item
              label="Name"
              name="name"
              rules={[{required: true, message: "Name is required"}]}
            >
              <Input disabled={modalMode === "edit"} placeholder="my-resource-quota" />
            </Form.Item>

            <Form.List name="hardEntries">
              {(fields, {add, remove}) => (
                <>
                  <div style={{marginBottom: 8, fontWeight: 500}}>Hard Limits</div>
                  {fields.map(({key, name, ...rest}) => (
                    <Space key={key} align="baseline" style={{display: "flex", marginBottom: 4}}>
                      <Form.Item
                        {...rest}
                        name={[name, "resource"]}
                        rules={[{required: true, message: "Resource required"}]}
                        style={{marginBottom: 0}}
                      >
                        <Select
                          placeholder="Resource"
                          style={{width: 220}}
                          options={resourceOptions}
                          showSearch
                          allowClear
                        />
                      </Form.Item>
                      <Form.Item
                        {...rest}
                        name={[name, "quantity"]}
                        rules={[{required: true, message: "Quantity required"}]}
                        style={{marginBottom: 0}}
                      >
                        <Input placeholder="e.g. 2, 4Gi, 500m" style={{width: 160}} />
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
                    Add Limit
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

export default ResourceQuotaListPage;
