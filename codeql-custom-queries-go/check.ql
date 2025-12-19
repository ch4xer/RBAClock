/**
 * @name Rbaclock
 * @kind test
 * @problem.severity warning
 */


import go

class CodeInjectionFlowConfig extends TaintTracking::Configuration {
    CodeInjectionFlowConfig() { this = "CodeInjectionFlowConfig" }

    override predicate isSource(DataFlow::Node source) { 
        source = any(DataFlow::ExprNode en | 
            isCR(en.asExpr().getType())
            // and en.getRoot().getChild(0).toString() = "renderLaunchManifest"
        )
        // source = any(DataFlow::ExprNode en)
    }

    override predicate isSink(DataFlow::Node sink) {
        // sink = any(DataFlow::ExprNode en| en.getRoot().getChild(0).toString() = "renderLaunchManifest" and (en.asExpr().getType().hasQualifiedName("k8s.io/api/core/v1", "Pod")))

        sink = any(DataFlow::ExprNode en |
            isInTargetBuiltin(en.asExpr()) 
            and not isFromCR(en)
            // and en.getRoot().getChild(0).toString() = "renderLaunchManifest"
        )
    }
}

// predicate belongsCR(Expr e) {
//     exists(|
//         e.(VariableName).getType().hasQualifiedName("atesting", "CustomResource") or 
//         e.(VariableName).getType().(PointerType).getBaseType().hasQualifiedName("atesting", "CustomResource") or
//         belongsCR(e.(SelectorExpr).getChild(0)) 
//     )
// }

predicate isTargetBuiltin(Type t) {
    exists( | 
        t.hasQualifiedName("k8s.io/api/core/v1", "Pod") 
        or t.hasQualifiedName("k8s.io/api/core/v1", "Service") 
        or t.hasQualifiedName("k8s.io/api/core/v1", "Secret") 
        or t.hasQualifiedName("k8s.io/api/core/v1", "Node") 
        or t.hasQualifiedName("k8s.io/api/core/v1", "ServiceAccount") 
        or t.hasQualifiedName("k8s.io/api/apps/v1", "Deployment")
        or t.hasQualifiedName("k8s.io/api/apps/v1", "ReplicaSet")
        or t.hasQualifiedName("k8s.io/api/apps/v1", "StatefulSet")
        or t.hasQualifiedName("k8s.io/api/apps/v1", "DaemonSet")
        or t.hasQualifiedName("k8s.io/api/batch/v1", "Job")
        or t.hasQualifiedName("k8s.io/api/batch/v1", "CronJob")
        or t.hasQualifiedName("k8s.io/api/batch/v1beta1", "CronJob")
        or t.hasQualifiedName("k8s.io/api/rbac/v1", "Role")
        or t.hasQualifiedName("k8s.io/api/rbac/v1", "RoleBinding")
        or t.hasQualifiedName("k8s.io/api/rbac/v1", "ClusterRole")
        or t.hasQualifiedName("k8s.io/api/rbac/v1", "ClusterRoleBinding")
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1", "MutatingWebhookConfiguration")
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1", "ValidatingWebhookConfiguration")
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1beta1", "MutatingWebhookConfiguration")
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1beta1", "ValidatingWebhookConfiguration")
        or t.hasQualifiedName("k8s.io/api/networking/v1", "NetworkPolicy")
        or t.hasQualifiedName("k8s.io/api/networking/v1", "Ingress")
        or t.hasQualifiedName("k8s.io/api/certificates/v1", "CertificateSigningRequest")
        or isTargetBuiltin(t.(PointerType).getBaseType())
    )
}

predicate isBuiltin(Type t) {
    exists( | 
        t.hasQualifiedName("k8s.io/api/core/v1", _) 
        or t.hasQualifiedName("k8s.io/api/apps/v1", _)
        or t.hasQualifiedName("k8s.io/api/batch/v1", _)
        or t.hasQualifiedName("k8s.io/api/batch/v1beta1", _)
        or t.hasQualifiedName("k8s.io/api/rbac/v1", _)
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1", _)
        or t.hasQualifiedName("k8s.io/api/admissionregistration/v1beta1", _)
        or t.hasQualifiedName("k8s.io/apimachinery/pkg/types", _)
        or t.hasQualifiedName("k8s.io/apimachinery/pkg/api/resource", _)
        or t.hasQualifiedName("k8s.io/apimachinery/pkg/apis/meta/v1", _)
        or t.hasQualifiedName("k8s.io/api/batch/v1", _)
        or t.hasQualifiedName("k8s.io/api/networking/v1", _)
        or t.hasQualifiedName("k8s.io/api/certificates/v1", _)
        or t.hasQualifiedName("k8s.io/api/policy/v1", _)
        or t.hasQualifiedName("k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1", _)
        or isBuiltin(t.(PointerType).getBaseType())
    )
}

// there may not be a VariableName, instead, StructLit
predicate isInTargetBuiltin(Expr e) {
    exists( | 
        // (e.getParent().(StructLit).getStructType().hasField("Spec", t) and t.getName() = "PodSpec") or
        // isInBuiltin(e.getParent())
        isTargetBuiltin(e.getType())
        or isInTargetBuiltin(e.getParent())
    )
}

predicate isCR(Type t) {
    exists(Field f1, Field f2|
        (
            t.getField(_).getType().getName() = "TypeMeta"
            and t.getField(_).getType().getName() = "ObjectMeta"
            and t.getField("Spec") = f1
            and t.getField("Status") = f2
            and not isBuiltin(t)
        ) or isCR(t.(PointerType).getBaseType())
    )
}


predicate hasCR(DataFlow::ExprNode en) {
    exists( | 
        // n.asExpr().getAChild().(VariableName).getType().hasQualifiedName("atesting", "CustomResource")
        isCR(en.asExpr().getType())
        or isCR(en.asExpr().getAChild().(Expr).getType())
    )
}

predicate isFromCR(DataFlow::Node en) {
    exists( |
        hasCR(en)
        or isFromCR(en.getAPredecessor())
    )
}

string getCRType(Expr e) {
    if e.getType() instanceof PointerType then result = e.getType().(PointerType).getBaseType().getQualifiedName() else result = e.getType().getQualifiedName()
}

string getBuiltinType(Expr e) {
    if isBuiltin(e.getType()) then result = e.getType().getQualifiedName() else result = getBuiltinType(e.getParent())
}

string getTargetBuiltinType(Expr e) {
    if isTargetBuiltin(e.getType()) then result = e.getType().getQualifiedName() else result = getTargetBuiltinType(e.getParent())
}

// how to get the map of CR -> Builtin => Taint Tracking 
// from DataFlow::ExprNode en, DataFlow::ExprNode source, DataFlow::ExprNode sink
// where 
// isFromCR(en)
// and isInBuiltin(en.asExpr()) 
// and not isBuiltin(en.asExpr().getType()) 
// and en.asExpr() instanceof VariableName
// and not en.asExpr() instanceof SelectorExpr 
// // and en.getRoot().getChild(0).toString() = "renderLaunchManifest"
// select en

from CodeInjectionFlowConfig config, DataFlow::Node source, DataFlow::Node sink
where
  config.hasFlow(source, sink)
  and source != sink
// select source, sink, source.asExpr().(VariableName).getType().getQualifiedName(), getBuiltinType(sink.asExpr())
select getCRType(source.asExpr()), getTargetBuiltinType(sink.asExpr())
// select source.asExpr().(VariableName).getType().getQualifiedName()

// can't go through the variable definition
// from DataFlow::ExprNode en
// where hasCR(en.getAPredecessor().getAPredecessor())
// select en.getAPredecessor().(DataFlow::Node)


// from DataFlow::Node source, DataFlow::Node sink, TrackingFlowConfig cfg
// where cfg.hasFlow(source, sink) and source != sink
// select source, sink, sink.getAPredecessor().getAPredecessor()


// from KeyValueExpr kv
// select kv.getParent().getParent().getParent().(StructLit).getTypeExpr()

// from DataFlow::ExprNode en, Type t
// where en.asExpr().getParent() instanceof KeyValueExpr and
// en.asExpr().getParent().getParent().getParent().getParent().getParent().getParent().getParent().(StructLit).getStructType().hasField("Spec", t) and t.getName() = "PodSpec"
// select en, en.getAPredecessor()

// from DataFlow::ExprNode en
// where belongsBuiltin(en.asExpr())
// // select en, en.getAPredecessor().getAPredecessor().asExpr().getAChild().(VariableName).getType().(StructType).getQualifiedName()
// select en, en.getAPredecessor().getAPredecessor()

