import React from "react";
import {
  Alert, Button, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Table, Tag, Tooltip
} from "antd";
import {DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined} from "@ant-design/icons";
import * as HPABackend from "./backend/HPABackend";
import * as NamespaceBackend from "./backend/NamespaceBackend";
import * as Setting from "./Setting";

const SCALE_TARGET_KINDS = ["Deployment", "StatefulSet", "ReplicaSet"];

class HPAListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hpas: [],
      namespaces: [],
      loading: true,
      error: null,
      modalVisible: false,
      modalMode: "add",
      submitting: false,
      editingHpa: null,
    };
    this.formRef = React.createRef();
  }

  componentDidMount() {
    this.fetchHPAs();
    this.fetchNamespaces();
  }

  fetchNamespaces() {
    NamespaceBackend.getNamespaces().then(res => {
      if (res.status === "ok") {
        this.setState({namespaces: res.data ?? []});
      }
    }).catch(() => {});
  }

  fetchHPAs() {
    this.setState({loading: true, error: null});
    HPABackend.getHPAs().then(res => {
      if (res.status === "ok") {
        this.setState({hpas: res.data ?? []});
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
    this.setState({modalVisible: true, modalMode: "add", editingHpa: null}, () => {
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({
          name: "",
          namespace: defaultNs,
          scaleTargetKind: "Deployment",
          scaleTargetName: "",
          minReplicas: 1,
          maxReplicas: 10,
          cpuTargetUtilization: 80,
        });
      }, 0);
    });
  }

  openEditModal(hpa) {
    this.setState({modalVisible: true, modalMode: "edit", editingHpa: hpa}, () => {
      setTimeout(() => {
        this.formRef.current?.setFieldsValue({
          name: hpa.name,
          namespace: hpa.namespace,
          scaleTargetKind: hpa.scaleTargetKind,
          scaleTargetName: hpa.scaleTargetName,
          minReplicas: hpa.minReplicas,
          maxReplicas: hpa.maxReplicas,
          cpuTargetUtilization: hpa.cpuTargetUtilization ?? null,
        });
      }, 0);
    });
  }

  closeModal() {
    this.setState({modalVisible: false, editingHpa: null});
  }

  handleSubmit() {
    this.formRef.current?.validateFields().then(values => {
      const payload = {
        name: values.name,
        namespace: values.namespace,
        scaleTargetKind: values.scaleTargetKind,
        scaleTargetName: values.scaleTargetName,
        minReplicas: values.minReplicas ?? 1,
        maxReplicas: values.maxReplicas,
        cpuTargetUtilization: values.cpuTargetUtilization ?? null,
      };

      this.setState({submitting: true});

      if (this.state.modalMode === "add") {
        HPABackend.addHPA(payload).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Horizontal Pod Autoscaler created");
            this.closeModal();
            this.fetchHPAs();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      } else {
        const hpa = this.state.editingHpa;
        HPABackend.updateHPA({...payload, resourceVersion: hpa.resourceVersion}).then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", "Horizontal Pod Autoscaler updated");
            this.closeModal();
            this.fetchHPAs();
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).catch(e => Setting.showMessage("error", e.message))
          .finally(() => this.setState({submitting: false}));
      }
    });
  }

  handleDelete(hpa) {
    HPABackend.deleteHPA(hpa.namespace, hpa.name).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", "Horizontal Pod Autoscaler deleted");
        this.fetchHPAs();
      } else {
        Setting.showMessage("error", res.msg);
      }
    }).catch(e => Setting.showMessage("error", e.message));
  }

  render() {
    const {hpas, namespaces, loading, error, modalVisible, modalMode, submitting} = this.state;

    const nsOptions = namespaces.map(ns => ({label: ns.name, value: ns.name}));

    const columns = [
      {title: "Namespace", dataIndex: "namespace", key: "namespace", width: 150},
      {title: "Name", dataIndex: "name", key: "name"},
      {
        title: "Scale Target",
        dataIndex: "scaleTargetRef",
        key: "scaleTargetRef",
        width: 200,
        render: v => <Tag>{v}</Tag>,
      },
      {title: "Min Replicas", dataIndex: "minReplicas", key: "minReplicas", width: 110},
      {title: "Max Replicas", dataIndex: "maxReplicas", key: "maxReplicas", width: 110},
      {
        title: "Current / Desired",
        key: "replicas",
        width: 140,
        render: (_, r) => `${r.currentReplicas} / ${r.desiredReplicas}`,
      },
      {
        title: "CPU Target %",
        dataIndex: "cpuTargetUtilization",
        key: "cpuTargetUtilization",
        width: 120,
        render: v => v !== null && v !== undefined ? `${v}%` : "-",
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
              title={`Delete Horizontal Pod Autoscaler "${record.name}"?`}
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
            message="Failed to fetch Horizontal Pod Autoscalers"
            description={error}
            style={{marginBottom: 16}}
            showIcon
          />
        )}

        <Table
          rowKey={r => `${r.namespace}/${r.name}`}
          columns={columns}
          dataSource={hpas}
          loading={loading}
          size="middle"
          scroll={{x: 1200}}
          pagination={{pageSize: 20}}
          locale={{emptyText: "No Horizontal Pod Autoscalers found"}}
          title={() => (
            <div>
              <span style={{fontWeight: 600}}>Horizontal Pod Autoscalers</span>
              &nbsp;&nbsp;&nbsp;&nbsp;
              <Button icon={<ReloadOutlined />} onClick={() => this.fetchHPAs()} loading={loading} size="small">
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
          title={modalMode === "add" ? "Add Horizontal Pod Autoscaler" : "Edit Horizontal Pod Autoscaler"}
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
              <Input disabled={modalMode === "edit"} placeholder="my-hpa" />
            </Form.Item>
            <Form.Item label="Scale Target Kind" name="scaleTargetKind" rules={[{required: true, message: "Scale target kind is required"}]}>
              <Select options={SCALE_TARGET_KINDS.map(k => ({label: k, value: k}))} />
            </Form.Item>
            <Form.Item label="Scale Target Name" name="scaleTargetName" rules={[{required: true, message: "Scale target name is required"}]}>
              <Input placeholder="my-deployment" />
            </Form.Item>
            <Space size="large" style={{width: "100%"}}>
              <Form.Item label="Min Replicas" name="minReplicas" rules={[{required: true, message: "Required"}]}>
                <InputNumber min={1} style={{width: 120}} />
              </Form.Item>
              <Form.Item label="Max Replicas" name="maxReplicas" rules={[{required: true, message: "Required"}]}>
                <InputNumber min={1} style={{width: 120}} />
              </Form.Item>
              <Form.Item
                label={
                  <Tooltip title="Target average CPU utilization (percentage). Leave empty to disable CPU-based scaling.">
                    CPU Target %
                  </Tooltip>
                }
                name="cpuTargetUtilization"
              >
                <InputNumber min={1} max={100} placeholder="80" style={{width: 100}} />
              </Form.Item>
            </Space>
          </Form>
        </Modal>
      </div>
    );
  }
}

export default HPAListPage;
